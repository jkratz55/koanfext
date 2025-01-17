package providers

import (
	"github.com/knadh/koanf/v2"
)

var _ koanf.Provider = (*ConfigMap)(nil)

type ConfigMap struct {
}

func ConfigMapProvider() *ConfigMap {
	return &ConfigMap{}
}

func (c *ConfigMap) ReadBytes() ([]byte, error) {
	// TODO implement me
	panic("implement me")
}

func (c *ConfigMap) Read() (map[string]interface{}, error) {
	// TODO implement me
	panic("implement me")
}

func (c *ConfigMap) Watch(cb func(event interface{}, err error)) error {
	return nil
}
