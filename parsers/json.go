package parsers

import (
	"encoding/json"

	"github.com/knadh/koanf/v2"
)

var _ koanf.Parser = (*JSON)(nil)

// JSON is a koanf.Parser for encoding/decoding JSON data.
type JSON struct{}

// JsonParser returns a koanf.Parser for JSON data
func JsonParser() *JSON {
	return &JSON{}
}

func (J *JSON) Unmarshal(bytes []byte) (map[string]interface{}, error) {
	content, err := ParseEnvironment(bytes)
	if err != nil {
		return nil, err
	}

	var out map[string]interface{}
	if err = json.Unmarshal(content, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (J *JSON) Marshal(m map[string]interface{}) ([]byte, error) {
	return json.Marshal(m)
}
