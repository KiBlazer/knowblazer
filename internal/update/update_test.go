package update

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestBuildAssetName(t *testing.T) {
	tests := []struct {
		tag    string
		goos   string
		goarch string
		want   string
	}{
		{"v0.2.0", "linux", "amd64", "knowblazer_0.2.0_linux_amd64.tar.gz"},
		{"0.2.0", "darwin", "arm64", "knowblazer_0.2.0_darwin_arm64.tar.gz"},
		{"v1.0.0", "windows", "amd64", "knowblazer_1.0.0_windows_amd64.zip"},
	}

	for _, tt := range tests {
		got, err := BuildAssetName(tt.tag, tt.goos, tt.goarch)
		if err != nil {
			t.Fatalf("BuildAssetName(%s, %s, %s) error = %v", tt.tag, tt.goos, tt.goarch, err)
		}
		if got != tt.want {
			t.Errorf("BuildAssetName(%s, %s, %s) = %q, want %q", tt.tag, tt.goos, tt.goarch, got, tt.want)
		}
	}
}

func createTarGz(filename string, content []byte) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	hdr := &tar.Header{
		Name: filename,
		Mode: 0o755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return nil, err
	}
	if _, err := tw.Write(content); err != nil {
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func createZip(filename string, content []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	w, err := zw.Create(filename)
	if err != nil {
		return nil, err
	}
	if _, err := w.Write(content); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func TestExtractBinaryTarGz(t *testing.T) {
	expected := []byte("#!/bin/sh\necho updated\n")
	archive, err := createTarGz("knowblazer", expected)
	if err != nil {
		t.Fatalf("createTarGz: %v", err)
	}

	extracted, err := ExtractBinary(archive, "knowblazer_0.2.0_linux_amd64.tar.gz", "linux")
	if err != nil {
		t.Fatalf("ExtractBinary() error = %v", err)
	}
	if !bytes.Equal(extracted, expected) {
		t.Errorf("ExtractBinary() got %q, want %q", extracted, expected)
	}
}

func TestExtractBinaryZip(t *testing.T) {
	expected := []byte("fake-windows-binary")
	archive, err := createZip("knowblazer.exe", expected)
	if err != nil {
		t.Fatalf("createZip: %v", err)
	}

	extracted, err := ExtractBinary(archive, "knowblazer_0.2.0_windows_amd64.zip", "windows")
	if err != nil {
		t.Fatalf("ExtractBinary() error = %v", err)
	}
	if !bytes.Equal(extracted, expected) {
		t.Errorf("ExtractBinary() got %q, want %q", extracted, expected)
	}
}

func TestRunSelfUpdateWorkflow(t *testing.T) {
	newBinaryContent := []byte("new-knowblazer-binary")
	tarBytes, err := createTarGz("knowblazer", newBinaryContent)
	if err != nil {
		t.Fatalf("createTarGz: %v", err)
	}

	assetName, _ := BuildAssetName("v99.0.0", runtime.GOOS, runtime.GOARCH)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/repos/KiBlazer/knowblazer/releases/latest":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"tag_name": "v99.0.0", "html_url": "https://github.com/KiBlazer/knowblazer/releases/tag/v99.0.0"}`)
		case r.URL.Path == "/KiBlazer/knowblazer/releases/download/v99.0.0/"+assetName:
			w.Header().Set("Content-Type", "application/gzip")
			w.Write(tarBytes)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	// Redirect github host to mock server via custom transport
	transport := &mockTransport{
		mockBase: ts.URL,
		wrapped:  http.DefaultTransport,
	}
	client := &http.Client{Transport: transport}

	targetDir := t.TempDir()
	targetBinary := filepath.Join(targetDir, "knowblazer")
	if err := os.WriteFile(targetBinary, []byte("old-binary"), 0o755); err != nil {
		t.Fatalf("write initial binary: %v", err)
	}

	cfg := Config{
		RepoOwner:    "KiBlazer",
		RepoName:     "knowblazer",
		HTTPClient:   client,
		TargetBinary: targetBinary,
		Force:        true,
	}

	var stdout bytes.Buffer
	if err := Run(context.Background(), cfg, &stdout); err != nil {
		t.Fatalf("Run() error = %v, output: %s", err, stdout.String())
	}

	updatedContent, err := os.ReadFile(targetBinary)
	if err != nil {
		t.Fatalf("read updated binary: %v", err)
	}
	if !bytes.Equal(updatedContent, newBinaryContent) {
		t.Fatalf("binary content not updated! got %q, want %q", updatedContent, newBinaryContent)
	}
}

type mockTransport struct {
	mockBase string
	wrapped  http.RoundTripper
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	cloned := req.Clone(req.Context())
	if req.URL.Host == "api.github.com" || req.URL.Host == "github.com" {
		mockURL, err := req.URL.Parse(m.mockBase + req.URL.Path)
		if err != nil {
			return nil, err
		}
		cloned.URL = mockURL
	}
	return m.wrapped.RoundTrip(cloned)
}
