package parsers

import (
	"github.com/knadh/koanf/v2"
	"github.com/pelletier/go-toml/v2"
)

var _ koanf.Parser = (*Toml)(nil)

type Toml struct{}

func TomlParser() *Toml {
	return &Toml{}
}

func (t *Toml) Unmarshal(bytes []byte) (map[string]interface{}, error) {
	content, err := ParseEnvironment(bytes)
	if err != nil {
		return nil, err
	}

	var out map[string]interface{}
	if err := toml.Unmarshal(content, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (t *Toml) Marshal(m map[string]interface{}) ([]byte, error) {
	return toml.Marshal(&m)
}
