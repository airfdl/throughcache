package throughcache

import (
	"time"

	"github.com/go-redis/redis"
)

type CommonRedis struct {
	client *redis.Client
	name   string
}

func NewCommonRedis(name string, client *redis.Client) CommonRedis {
	return CommonRedis{
		name:   name,
		client: client,
	}
}

func (r CommonRedis) Name() string {
	return r.name
}

func (r CommonRedis) MGet(keys []string) (map[string][]byte, error) {
	pipeline := r.client.Pipeline()
	defer func() { _ = pipeline.Close() }()
	for _, key := range keys {
		pipeline.Get(key)
	}
	datas, err := pipeline.Exec()
	if err != nil {
		return nil, err
	}
	ret := make(map[string][]byte)
	for index, value := range datas {
		v, err := value.(*redis.StringCmd).Bytes()
		if err != nil {
			continue
		}
		key := keys[index]
		ret[key] = v
	}
	return ret, nil
}

func (r CommonRedis) Expire(keys []string, expire time.Duration) error {
	pipeline := r.client.Pipeline()
	defer func() { _ = pipeline.Close() }()
	for _, key := range keys {
		pipeline.Expire(key, expire)
	}
	_, err := pipeline.Exec()
	return err
}

func (r CommonRedis) MSet(data map[string][]byte, expire time.Duration) error {
	pipeline := r.client.Pipeline()
	for key, value := range data {
		if value == nil {
			continue
		}
		if expire >= 0 {
			pipeline.Set(key, value, expire)
		}
	}
	defer func() { _ = pipeline.Close() }()
	_, err := pipeline.Exec()
	return err
}

func (r CommonRedis) DeleteKeys(keys []string) error {
	pipeline := r.client.Pipeline()
	pipeline.Del(keys...)
	defer func() { _ = pipeline.Close() }()
	_, err := pipeline.Exec()
	return err
}
