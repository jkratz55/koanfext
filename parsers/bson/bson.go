package bson

import (
	"github.com/knadh/koanf/v2"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/jkratz55/koanfext/parsers/env"
)

var _ koanf.Parser = (*BSON)(nil)

// BSON is a koanf.Parser for encoding/decoding BSON data.
type BSON struct{}

// Parser returns a koanf.Parser for BSON data
func Parser() *BSON {
	return &BSON{}
}

func (B BSON) Unmarshal(bytes []byte) (map[string]interface{}, error) {
	content, err := env.ParseEnvironment(bytes)
	if err != nil {
		return nil, err
	}

	var out map[string]interface{}
	if err = bson.Unmarshal(content, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (B BSON) Marshal(m map[string]interface{}) ([]byte, error) {
	return bson.Marshal(m)
}
