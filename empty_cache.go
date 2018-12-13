package throughcache

import "time"

type emptyCache struct{}

var emptyCacher = new(emptyCache)

func CacheTODO() Cacher {
	return emptyCacher
}

func (emptyCache) MGet([]string) (map[string][]byte, error) {
	return nil, nil
}
func (emptyCache) MSet(map[string][]byte, time.Duration) error {
	return nil
}
func (emptyCache) DeleteKeys([]string) error {
	return nil
}
func (emptyCache) Expire([]string, time.Duration) error {
	return nil
}
func (emptyCache) Name() string {
	return "emptyCache"
}
