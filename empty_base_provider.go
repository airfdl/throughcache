package throughcache

import (
	"context"
	"encoding/json"
	"errors"
)

type emptyBase struct{}

var (
	todoBase     = new(emptyBase)
	ErrEmptyBase = errors.New("EmptyBase")
)

func BaseTODO() BaseDataProvider {
	return todoBase
}

func (emptyBase) DataProvider(context.Context, []Queryer) ([]Modeler, error) {
	return nil, ErrEmptyBase
}

func (emptyBase) Marshal(Modeler) ([]byte, error) {
	return nil, ErrEmptyBase
}

func (emptyBase) Unmarshal(IDer, []byte) (Modeler, error) {
	return nil, ErrEmptyBase
}

func (emptyBase) Name() string {
	return "EmptyBase"
}

var JsonCoder jsonCode

type jsonCode struct{}

func (jsonCode) Marshal(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func (jsonCode) Unmarshal(byteValue []byte, v interface{}) error {
	return json.Unmarshal(byteValue, v)
}
