package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/knowblazer/knowblazer/internal/version"
)

const (
	DefaultRepoOwner = "KiBlazer"
	DefaultRepoName  = "knowblazer"
)

type Config struct {
	RepoOwner    string
	RepoName     string
	HTTPClient   *http.Client
	TargetBinary string // path to current binary to replace (defaults to os.Executable())
	Force        bool
}

type ReleaseInfo struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
}

// CheckLatestTag returns the latest release tag (e.g. "v0.1.0").
func CheckLatestTag(ctx context.Context, client *http.Client, owner, repo string) (string, error) {
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "knowblazer-update/"+version.Get())
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		var rel ReleaseInfo
		if err := json.NewDecoder(resp.Body).Decode(&rel); err == nil && rel.TagName != "" {
			return rel.TagName, nil
		}
	}
	if resp != nil {
		_ = resp.Body.Close()
	}

	// Fallback to checking the redirect location of /releases/latest
	latestURL := fmt.Sprintf("https://github.com/%s/%s/releases/latest", owner, repo)
	checkReq, err := http.NewRequestWithContext(ctx, http.MethodHead, latestURL, nil)
	if err != nil {
		return "", err
	}
	checkReq.Header.Set("User-Agent", "knowblazer-update/"+version.Get())

	// Create a transport that doesn't follow redirects to capture the Location header
	noRedirectClient := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	checkResp, err := noRedirectClient.Do(checkReq)
	if err == nil {
		defer checkResp.Body.Close()
		loc := checkResp.Header.Get("Location")
		if loc != "" {
			parts := strings.Split(loc, "/")
			if len(parts) > 0 {
				tag := parts[len(parts)-1]
				if strings.HasPrefix(tag, "v") {
					return tag, nil
				}
			}
		}
	}

	return "", fmt.Errorf("could not determine latest release for %s/%s", owner, repo)
}

// BuildAssetName generates the goreleaser asset filename for the platform.
func BuildAssetName(tag, goos, goarch string) (string, error) {
	cleanVersion := strings.TrimPrefix(tag, "v")
	switch goos {
	case "windows":
		return fmt.Sprintf("knowblazer_%s_%s_%s.zip", cleanVersion, goos, goarch), nil
	case "linux", "darwin":
		return fmt.Sprintf("knowblazer_%s_%s_%s.tar.gz", cleanVersion, goos, goarch), nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", goos)
	}
}

// Run performs the update workflow.
func Run(ctx context.Context, cfg Config, stdout io.Writer) error {
	if cfg.RepoOwner == "" {
		cfg.RepoOwner = DefaultRepoOwner
	}
	if cfg.RepoName == "" {
		cfg.RepoName = DefaultRepoName
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 60 * time.Second}
	}

	current := version.Get()
	fmt.Fprintf(stdout, "Checking for latest release of knowblazer (current: v%s)...\n", current)

	latestTag, err := CheckLatestTag(ctx, cfg.HTTPClient, cfg.RepoOwner, cfg.RepoName)
	if err != nil {
		return fmt.Errorf("check update failed: %w", err)
	}

	latestClean := strings.TrimPrefix(latestTag, "v")
	if !cfg.Force && (current == latestClean || strings.HasPrefix(current, latestClean+"-") || strings.HasPrefix(current, latestClean+".")) {
		fmt.Fprintf(stdout, "knowblazer is already up to date (current: v%s, latest: %s).\n", current, latestTag)
		return nil
	}

	fmt.Fprintf(stdout, "Found new version %s. Downloading...\n", latestTag)

	assetName, err := BuildAssetName(latestTag, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}

	downloadURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s",
		cfg.RepoOwner, cfg.RepoName, latestTag, assetName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "knowblazer-update/"+current)

	resp, err := cfg.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned HTTP %d from %s", resp.StatusCode, downloadURL)
	}

	archiveBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read archive: %w", err)
	}

	binaryBytes, err := ExtractBinary(archiveBytes, assetName, runtime.GOOS)
	if err != nil {
		return fmt.Errorf("extract binary: %w", err)
	}

	target := cfg.TargetBinary
	if target == "" {
		target, err = os.Executable()
		if err != nil {
			return fmt.Errorf("find executable path: %w", err)
		}
		target, err = filepath.EvalSymlinks(target)
		if err != nil {
			return fmt.Errorf("eval symlink: %w", err)
		}
	}

	if err := applyBinary(binaryBytes, target); err != nil {
		return err
	}

	fmt.Fprintf(stdout, "Successfully updated knowblazer to %s (%s)!\n", latestTag, target)
	return nil
}

// ExtractBinary extracts the knowblazer executable from an in-memory tar.gz or zip archive.
func ExtractBinary(data []byte, filename string, goos string) ([]byte, error) {
	expectedBinary := "knowblazer"
	if goos == "windows" {
		expectedBinary = "knowblazer.exe"
	}

	if strings.HasSuffix(filename, ".tar.gz") || strings.HasSuffix(filename, ".tgz") {
		gr, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
		defer gr.Close()

		tr := tar.NewReader(gr)
		for {
			header, err := tr.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			if filepath.Base(header.Name) == expectedBinary {
				return io.ReadAll(tr)
			}
		}
		return nil, fmt.Errorf("binary %q not found in archive", expectedBinary)
	}

	if strings.HasSuffix(filename, ".zip") {
		zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return nil, err
		}
		for _, f := range zr.File {
			if filepath.Base(f.Name) == expectedBinary {
				rc, err := f.Open()
				if err != nil {
					return nil, err
				}
				defer rc.Close()
				return io.ReadAll(rc)
			}
		}
		return nil, fmt.Errorf("binary %q not found in zip archive", expectedBinary)
	}

	return nil, errors.New("unsupported archive format")
}

func applyBinary(data []byte, targetPath string) error {
	dir := filepath.Dir(targetPath)

	// Test if directory is writable
	tempFile, err := os.CreateTemp(dir, "knowblazer-update-*.tmp")
	if err != nil {
		if os.IsPermission(err) {
			return fmt.Errorf("permission denied updating %s. Try running with elevated permissions (e.g. sudo knowblazer update)", targetPath)
		}
		return fmt.Errorf("create temporary update file: %w", err)
	}
	tempName := tempFile.Name()
	defer os.Remove(tempName)

	if _, err := tempFile.Write(data); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("write binary: %w", err)
	}
	if err := tempFile.Chmod(0o755); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("chmod binary: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	// On Windows, running binaries cannot be directly overwritten with rename
	if runtime.GOOS == "windows" {
		oldFile := targetPath + ".old"
		_ = os.Remove(oldFile)
		if err := os.Rename(targetPath, oldFile); err != nil {
			return fmt.Errorf("backup current binary on windows: %w", err)
		}
	}

	if err := os.Rename(tempName, targetPath); err != nil {
		return fmt.Errorf("replace binary at %s: %w", targetPath, err)
	}

	return nil
}
