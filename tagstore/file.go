package tagstore

import (
	"context"
	"io/ioutil"
	"path"
)

// File is an implementation of TagStore using the local file system, rooted at the specified
// directory.
type File struct {
	DirName string
}

func (s File) Set(ctx context.Context, name string, value []byte) error {
	return ioutil.WriteFile(path.Join(s.DirName, name), value, 0644)
}

func (s File) Get(ctx context.Context, name string) ([]byte, error) {
	return ioutil.ReadFile(path.Join(s.DirName, name))
}

func (s File) List(ctx context.Context) ([]string, error) {
	files, err := ioutil.ReadDir(s.DirName)
	if err != nil {
		return nil, err
	}

	fileNames := []string{}
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}

	return fileNames, nil
}
