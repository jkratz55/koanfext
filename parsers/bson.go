package parsers

import (
	"github.com/knadh/koanf/v2"
	"go.mongodb.org/mongo-driver/bson"
)

var _ koanf.Parser = (*BSON)(nil)

type BSON struct{}

func BsonParser() *BSON {
	return &BSON{}
}

func (B BSON) Unmarshal(bytes []byte) (map[string]interface{}, error) {
	content, err := ParseEnvironment(bytes)
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
