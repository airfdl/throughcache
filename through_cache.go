package throughcache

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"time"
)

var (
	emptyValue = []byte("nil")
	asyncer    = NewAsyncQue(10000, 2*runtime.NumCPU())
)

type ThroughCache struct {
	name   string
	cacher Cacher
	baser  BaseDataProvider
	*Options
}

func NewThroughCache(name string, baser BaseDataProvider, cacher Cacher, opts ...Option) *ThroughCache {
	options := new(Options)
	for _, opt := range opts {
		opt.f(options)
	}
	t := &ThroughCache{
		name:    name,
		cacher:  cacher,
		baser:   baser,
		Options: options,
	}
	return t
}

func (t *ThroughCache) MGetValue(ctx context.Context, querys []Queryer) (map[string]Modeler, error) {
	querys, err := t.checkGids(querys)
	if err != nil {
		return nil, err
	}
	id2Modeler, missQuerys, err := t.MGetCache(ctx, querys)
	if err != nil && !t.mustToBaser() {
		logs.Error("ThroughCache:%v MGetCache error:%v", t.Name(), err)
		return id2Modeler, err
	}
	if len(missQuerys) == 0 {
		return id2Modeler, err
	}
	baseId2Modeler, err := t.MGetBase(ctx, missQuerys)
	if err != nil {
		if len(id2Modeler) != 0 {
			return id2Modeler, nil
		} else {
			return nil, err
		}
	}
	for id, modeler := range baseId2Modeler {
		id2Modeler[id] = modeler
	}
	logs.Info("ThroughCache:%v after base have count:%v", t.Name(), len(id2Modeler))
	if err == nil && (t.isSetCache() || t.isSetEmptyValueCache()) {
		missId2Query := Id2Queryer(missQuerys)
		if t.syncSetCache() {
			t.SetCache(ctx, missId2Query, baseId2Modeler)
		} else {
			t.AsyncSetCache(ctx, missId2Query, baseId2Modeler)
		}
	}
	return id2Modeler, nil
}

func (t *ThroughCache) AsyncUpdateCache(ctx context.Context, querys []Queryer) {
	if len(querys) == 0 {
		return
	}
	t.asyncRun(ctx, func() { t.UpdataCache(ctx, querys) })
}

func (t *ThroughCache) asyncRun(ctx context.Context, f func()) {
	err := asyncer.Send(f)
	if err != nil {
		logs.Error("send async error:%v", err)
		if !t.asyncFailBack() {
			return
		}
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	done := make(chan struct{})
	go func() {
		defer func() { close(done) }()
		f()
	}()

	select {
	case <-ctx.Done():
	case <-done:
	}
}

func (t *ThroughCache) UpdataCache(ctx context.Context, querys []Queryer) {
	if len(querys) == 0 {
		return
	}
	id2Modeler, err := t.MGetBase(ctx, querys)
	if err != nil {
		return
	}
	id2Query := Id2Queryer(querys)
	t.SetCache(ctx, id2Query, id2Modeler)
}

func (t *ThroughCache) AsyncSetCacheV2(ctx context.Context, missQuerys []Queryer, modelers []Modeler) {
	if len(missQuerys) == 0 {
		return
	}
	t.asyncRun(ctx, func() { t.SetCacheV2(ctx, missQuerys, modelers) })
}

func (t *ThroughCache) AsyncSetCache(ctx context.Context, missId2Query map[string]Queryer, id2Modeler map[string]Modeler) {
	if len(missId2Query) == 0 {
		return
	}
	t.asyncRun(ctx, func() { t.SetCache(ctx, missId2Query, id2Modeler) })
}

func (t *ThroughCache) DeleteKeys(ctx context.Context, querys ...Queryer) error {
	keys := Keys(querys)
	logs.Info("ThroughCache:%v keys:%s", t.Name(), keys)
	if len(keys) == 0 {
		return nil
	}
	return t.cacher.DeleteKeys(keys)
}

func (t *ThroughCache) ExpireKeys(ctx context.Context, expire time.Duration, querys ...Queryer) error {
	keys := Keys(querys)
	logs.Info("ThroughCache:%v keys:%s", t.Name(), keys)
	if len(keys) == 0 {
		return nil
	}
	return t.cacher.Expire(keys, expire)
}

func (t *ThroughCache) SetCacheV2(ctx context.Context, missQuerys []Queryer, modelers []Modeler) {
	if len(missQuerys) == 0 {
		return
	}
	_, _, missId2Query := QueryerRelations(missQuerys)
	id2Modeler := make(map[string]Modeler)
	for idx, _ := range modelers {
		id2Modeler[modelers[idx].ID()] = modelers[idx]
	}
	t.SetCache(ctx, missId2Query, id2Modeler)
}

func (t *ThroughCache) SetCache(ctx context.Context, missId2Query map[string]Queryer, id2Modeler map[string]Modeler) {
	if len(missId2Query) == 0 {
		return
	}
	missData := make(map[string][]byte)
	missEmptyData := make(map[string][]byte)
	emptyValueQueryIds := make([]string, 0, len(missId2Query))
	normalValueQueryIds := make([]string, 0, len(missId2Query))
	for id, query := range missId2Query {
		modeler, found := id2Modeler[id]
		if !found && t.isSetEmptyValueCache() {
			//value 给缓存默认值
			missEmptyData[query.MakeKey()] = getEmptyValue()
			emptyValueQueryIds = append(emptyValueQueryIds, query.ID())
		} else {
			normalValueQueryIds = append(normalValueQueryIds, query.ID())
			if t.isSetCache() {
				byteValue, err := t.baser.Marshal(modeler)
				if err != nil {
					logs.Error("ThroughCache %v call marshal %v error %v", t.Name(), t.baser.Name(), err)
					continue
				}
				missData[query.MakeKey()] = byteValue
			}
		}
	}
	logs.Info("Can Set NormalQueryIds:%v, EmptyQueryIds:%v", normalValueQueryIds, emptyValueQueryIds)
	if len(missData) > 0 && t.isSetCache() {
		err := t.cacher.MSet(missData, t.expireTime())
		if err != nil {
			logs.Error("ThroughCache %v call cache set error %v", t.Name(), err)
		}
	}
	if len(missEmptyData) > 0 && t.isSetEmptyValueCache() {
		_ = t.cacher.MSet(missEmptyData, t.emptyValueExpireTime())
	}
}

func (t *ThroughCache) MGetCache(ctx context.Context, querys []Queryer) (map[string]Modeler, []Queryer, error) {
	if len(querys) == 0 {
		return nil, nil, nil
	}
	startTime := time.Now()
	id2Modeler, missQuerys, err := t.mGetCache(ctx, querys)
	cacheSpendTime := ElapsedMillisecondsStr(startTime)
	logs.Info("ThroughCache:%v MGetCache HitGidCount:%v missGidCount:%v MissGids:%v SpendTime:%vms Error:%v", t.Name(), len(id2Modeler), len(missQuerys), IDs(missQuerys), cacheSpendTime, err)
	return id2Modeler, missQuerys, err
}

func (t *ThroughCache) mGetCache(ctx context.Context, querys []Queryer) (map[string]Modeler, []Queryer, error) {
	keys := Keys(querys)
	if len(keys) == 0 {
		return nil, nil, nil
	}
	cacheData, err := t.cacher.MGet(keys)
	if err != nil {
		logs.Error("ThroughCache %v call cache %v error %v", t.Name(), t.baser.Name(), err)
		return nil, querys, err
	}
	missId2Query := make(map[string]Queryer)
	missQuerys := make([]Queryer, 0, 0)
	cacheID2Modeler := make(map[string]Modeler)
	emptyValueQuerys := make([]Queryer, 0, len(cacheData))
	for _, query := range querys {
		byteValue, found := cacheData[query.MakeKey()]
		if !found {
			if _, found := missId2Query[query.ID()]; !found {
				missQuerys = append(missQuerys, query)
				missId2Query[query.ID()] = query
			}
		} else {
			if isEmptyValue(byteValue) {
				//一个无效的默认值
				emptyValueQuerys = append(emptyValueQuerys, query)
				continue
			}
			logs.Info("ThroughCache:%v marshal key:%v byteValue len:%v", t.Name(), query.MakeKey(), len(byteValue))
			modeler, err := t.baser.Unmarshal(query, byteValue)
			if err != nil {
				logs.Error("ThroughCache %v call unmarshal %v error %v", t.Name(), t.baser.Name(), err)
				continue
			}
			cacheID2Modeler[query.ID()] = modeler
		}
	}
	logs.Info("%v get cache emptyValueGids:%v", t.Name(), IDs(emptyValueQuerys))
	return cacheID2Modeler, missQuerys, err
}

func (t *ThroughCache) MGetBase(ctx context.Context, querys []Queryer) (baseId2Modeler map[string]Modeler, err error) {
	startTime := time.Now()
	if len(querys) > 0 {
		baseId2Modeler, err = t.MGetBase(ctx, querys)
	}
	baseSpendTime := ElapsedMillisecondsStr(startTime)
	logs.Info("ThroughCache:%v MGetBase SpendTime:%v", t.Name(), baseSpendTime)
	return baseId2Modeler, err
}

func (t *ThroughCache) mGetBase(ctx context.Context, querys []Queryer) (map[string]Modeler, error) {
	if len(querys) == 0 {
		return nil, nil
	}
	startTime := time.Now()
	modelers, err := t.baser.DataProvider(ctx, querys)
	elapsed := ElapsedMilliseconds(startTime)
	_ = metricsClient.EmitTimer(fmt.Sprintf("call.%v.latency", t.baser.Name()), elapsed, t.tagkv())
	_ = metricsClient.EmitCounter(fmt.Sprintf("call.%v.num", t.baser.Name()), 1, t.tagkv())
	if err != nil {
		_ = metricsClient.EmitCounter(fmt.Sprintf("call.%v.error", t.baser.Name()), 1, t.tagkv())
		logs.Error("ThroughCache %v call base %v error %v", t.Name(), t.baser.Name(), err)
		return nil, err
	}
	id2Modeler := make(map[string]Modeler)
	for idx, _ := range modelers {
		id2Modeler[modelers[idx].ID()] = modelers[idx]
	}
	return id2Modeler, err
}

func (t *ThroughCache) checkGids(querys []Queryer) ([]Queryer, error) {
	out := make([]Queryer, 0, len(querys))
	for _, query := range querys {
		if reflect.ValueOf(query).IsNil() {
			continue
		}
		out = append(out, query)
	}
	if len(out) == 0 {
		return out, errors.New("empty querys")
	}
	return out, nil
}

func (t *ThroughCache) tagkv() map[string]string {
	return map[string]string{
		"method": t.Name(),
	}
}

func (t ThroughCache) Name() string {
	return t.name
}

func isEmptyValue(v []byte) bool {
	if len(v) == len(emptyValue) && string(v) == string(emptyValue) {
		return true
	}
	return false
}

func getEmptyValue() []byte {
	return emptyValue
}
