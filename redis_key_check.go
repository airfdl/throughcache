package throughcache

import (
	"sync"
)

var cacheKeyMap = make(map[string]map[string]struct{})
var keyMapLock sync.Mutex

func CacheKeyChecker(cacheName, prefix string) bool {
	//预检查合适
	keyMapLock.Lock()
	defer keyMapLock.Unlock()
	var isOk bool
	keyMap, find := cacheKeyMap[cacheName]
	if !find {
		keyMap = make(map[string]struct{})
		cacheKeyMap[cacheName] = keyMap
	}
	_, found := keyMap[prefix]
	if found {
		logs.Error("cache:%v already have prefix:%v", cacheName, prefix)
	} else {
		keyMap[prefix] = struct{}{}
		isOk = true
	}
	//ShowCacheKeyMap()
	return isOk
}

func ShowCacheKeyMap() {
	logs.Info("==================cacheKey check start======================")
	for cacheName, keyMap := range cacheKeyMap {
		for key, _ := range keyMap {
			logs.Info("cacheName:%s key:[%s]", cacheName, key)
		}
	}
	logs.Info("==================cacheKey check end======================")
}
