package throughcache

import (
	"context"
	"time"
)

type Cacher interface {
	MGet([]string) (map[string][]byte, error)
	MSet(map[string][]byte, time.Duration) error
	DeleteKeys([]string) error
	Expire([]string, time.Duration) error
	Name() string
}

type BaseDataProvider interface {
	DataProvider(context.Context, []Queryer) ([]Modeler, error)
	Name() string
	Serializer
}

type Serializer interface {
	Marshal(Modeler) ([]byte, error)
	Unmarshal(IDer, []byte) (Modeler, error)
}

type Modeler interface {
	IDer
}

type Queryer interface {
	Keyer
	IDer
}

type Keyer interface {
	//save cache key
	MakeKey() string
}

type IDer interface {
	ID() string
}
