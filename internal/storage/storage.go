package storage

import (
	"io"
	"os"
	"path/filepath"
)

type Storage interface {
	Save(filename string, data io.Reader) (string, error)
	Load(filename string) ([]byte, error)
	Delete(filename string) error
}

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) *LocalStorage {
	os.MkdirAll(basePath, 0755)
	return &LocalStorage{basePath: basePath}
}

func (s *LocalStorage) Save(filename string, data io.Reader) (string, error) {
	fullPath := filepath.Join(s.basePath, filename)
	out, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer out.Close()
	_, err = io.Copy(out, data)
	if err != nil {
		return "", err
	}
	return fullPath, nil
}

func (s *LocalStorage) Load(filename string) ([]byte, error) {
	fullPath := filepath.Join(s.basePath, filename)
	return os.ReadFile(fullPath)
}

func (s *LocalStorage) Delete(filename string) error {
	fullPath := filepath.Join(s.basePath, filename)
	return os.Remove(fullPath)
}
