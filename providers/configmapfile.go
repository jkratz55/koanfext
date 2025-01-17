package providers

import (
	"github.com/knadh/koanf/v2"
)

var _ koanf.Provider = (*ConfigMapFile)(nil)

type ConfigMapFile struct {
}

func ConfigMapFileProvider() *ConfigMapFile {
	return &ConfigMapFile{}
}

func (c *ConfigMapFile) ReadBytes() ([]byte, error) {
	// TODO implement me
	panic("implement me")
}

func (c *ConfigMapFile) Read() (map[string]interface{}, error) {
	// TODO implement me
	panic("implement me")
}

func (c *ConfigMapFile) Watch(cb func(event interface{}, err error)) error {
	return nil
}
