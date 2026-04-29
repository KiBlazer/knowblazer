package backup

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const encryptedMagic = "KNOWBLAZER-BACKUP-1\n"

func Create(repoRoot string, output string, passphrase string) error {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	err := filepath.WalkDir(repoRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil || rel == "." {
			return err
		}
		if strings.HasPrefix(rel, ".git") {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(rel)
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(tw, file)
		return err
	})
	if err != nil {
		return err
	}
	if err := tw.Close(); err != nil {
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}
	data := buf.Bytes()
	if passphrase != "" {
		var err error
		data, err = encrypt(data, passphrase)
		if err != nil {
			return err
		}
	}
	return os.WriteFile(output, data, 0o600)
}

func Restore(input string, target string, passphrase string) error {
	if entries, err := os.ReadDir(target); err == nil && len(entries) > 0 {
		return fmt.Errorf("restore target is non-empty")
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	data, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	if bytes.HasPrefix(data, []byte(encryptedMagic)) {
		data, err = decrypt(data, passphrase)
		if err != nil {
			return err
		}
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		return err
	}
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		clean := filepath.Clean(header.Name)
		if filepath.IsAbs(clean) || strings.HasPrefix(clean, "..") {
			return fmt.Errorf("unsafe backup path: %s", header.Name)
		}
		path := filepath.Join(target, clean)
		if header.FileInfo().IsDir() {
			if err := os.MkdirAll(path, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return err
		}
		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		if _, err := io.Copy(file, tr); err != nil {
			file.Close()
			return err
		}
		if err := file.Close(); err != nil {
			return err
		}
	}
	return nil
}

func encrypt(data []byte, passphrase string) ([]byte, error) {
	block, err := aes.NewCipher(key(passphrase))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	out := append([]byte(encryptedMagic), nonce...)
	out = gcm.Seal(out, nonce, data, nil)
	return out, nil
}

func decrypt(data []byte, passphrase string) ([]byte, error) {
	if passphrase == "" {
		return nil, fmt.Errorf("passphrase is required for encrypted backup")
	}
	block, err := aes.NewCipher(key(passphrase))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	data = data[len(encryptedMagic):]
	if len(data) < gcm.NonceSize() {
		return nil, fmt.Errorf("encrypted backup is truncated")
	}
	nonce := data[:gcm.NonceSize()]
	ciphertext := data[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func key(passphrase string) []byte {
	sum := sha256.Sum256([]byte(passphrase))
	return sum[:]
}
