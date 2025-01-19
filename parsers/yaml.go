package parsers

import (
	"github.com/knadh/koanf/v2"
	"gopkg.in/yaml.v3"
)

var _ koanf.Parser = (*Yaml)(nil)

// Yaml is a koanf.Parser for encoding/decoding yaml data.
type Yaml struct{}

// YamlParser returns a koanf.Parser for YAML data
func YamlParser() *Yaml {
	return &Yaml{}
}

func (y *Yaml) Unmarshal(bytes []byte) (map[string]interface{}, error) {
	content, err := ParseEnvironment(bytes)
	if err != nil {
		return nil, err
	}

	var out map[string]interface{}
	if err := yaml.Unmarshal(content, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (y *Yaml) Marshal(m map[string]interface{}) ([]byte, error) {
	return yaml.Marshal(m)
}
