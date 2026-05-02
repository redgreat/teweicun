/**
 * 功能：local.go
 * 创建时间：2026-04-18
 * 创建人：wangcw
 */

package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) (*LocalStorage, error) {
	// Create base path if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}
	return &LocalStorage{basePath: basePath}, nil
}

func (s *LocalStorage) Upload(filename string, content io.Reader) (string, error) {
	targetPath := filepath.Join(s.basePath, filename)
	
	// Create subdirectories if necessary
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return "", err
	}

	out, err := os.Create(targetPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, content)
	if err != nil {
		return "", err
	}

	return filename, nil
}

func (s *LocalStorage) GetURL(filename string) (string, error) {
	// For local storage, URL might just be a relative path or a full path
	// In a real app, you might serve files via /uploads route
	return fmt.Sprintf("/uploads/%s", filename), nil
}

func (s *LocalStorage) Delete(filename string) error {
	targetPath := filepath.Join(s.basePath, filename)
	return os.Remove(targetPath)
}
