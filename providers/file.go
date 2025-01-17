package providers

import (
	"fmt"
	"os"

	"github.com/knadh/koanf/v2"
)

var _ koanf.Provider = (*File)(nil)

type File struct {
	path string
}

func FileProvider(path string) *File {
	return &File{
		path: path,
	}
}

func (f *File) ReadBytes() ([]byte, error) {
	return os.ReadFile(f.path)
}

func (f *File) Read() (map[string]interface{}, error) {
	return nil, fmt.Errorf("%T does not support Read()", f)
}

func (f *File) Watch(cb func(event interface{}, err error)) error {
	return nil
}
