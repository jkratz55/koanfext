package toml

import (
	"github.com/knadh/koanf/v2"
	"github.com/pelletier/go-toml/v2"

	"github.com/jkratz55/koanfext/parsers/env"
)

var _ koanf.Parser = (*Toml)(nil)

// Toml is a koanf.Parser for encoding/decoding Toml files.
type Toml struct{}

// Parser returns a koanf.Parser for TOML data
func Parser() *Toml {
	return &Toml{}
}

func (t *Toml) Unmarshal(bytes []byte) (map[string]interface{}, error) {
	content, err := env.ParseEnvironment(bytes)
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
