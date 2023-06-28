package release

import (
	"github.com/otiai10/copy"
	"os"
	"path"
)

type FileSystem struct {
	target string
}

func NewFileSystem(target string) *FileSystem {
	return &FileSystem{target: target}
}

func (f *FileSystem) UploadFile(from string, to string) error {
	return copy.Copy(from, path.Join(f.target, to))
}

func (f *FileSystem) DownloadContent(from string) (string, error) {
	file, err := os.ReadFile(path.Join(f.target, from))
	if err != nil {
		return "", err
	}
	return string(file), nil
}

func (f *FileSystem) UploadContent(content string, to string) error {
	return os.WriteFile(path.Join(f.target, to), []byte(content), 0644)
}
