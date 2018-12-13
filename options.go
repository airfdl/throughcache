package throughcache

import (
	"time"
)

func NewDefaultOptions(expireTime, emptyValueExpireTime time.Duration) []Option {
	return []Option{
		WithExpireTimeout(expireTime),
		WithEmptyValueExpireTimeout(emptyValueExpireTime),
		WithSetEmptyValueCahe(),
		WithSetCache(),
		WithAsyncFailbackSync(),
	}
}

type Option struct {
	f func(*Options)
}

const (
	setCache int64 = 1 << iota
	setEmptyValueCache
	syncSetCache
	mustToBase
	asyncFailBack
)

func WithAsyncFailbackSync() Option {
	//异步失败时候直接同步
	return Option{f: func(o *Options) { o.bit1 |= asyncFailBack }}
}

func WithExpireTimeout(expireTime time.Duration) Option {
	return Option{f: func(o *Options) { o.expireDuration = expireTime }}
}

func WithEmptyValueExpireTimeout(expireTime time.Duration) Option {
	if expireTime <= 0 {
		expireTime = time.Minute * 30
	}
	return Option{f: func(o *Options) { o.emptyValueExpireDuration = expireTime }}
}

func WithSetEmptyValueCahe() Option {
	return Option{f: func(o *Options) { o.bit1 |= setEmptyValueCache }}
}

func WithSetCache() Option {
	return Option{f: func(o *Options) { o.bit1 |= setCache }}
}

func WithSyncSetCache() Option {
	return Option{f: func(o *Options) { o.bit1 |= syncSetCache }}
}

func WithMustToBaser() Option {
	//cache failed still to base data porvider, u must ensure baser
	//can still work with long cache problem
	return Option{f: func(o *Options) { o.bit1 |= mustToBase }}
}

type Options struct {
	expireDuration           time.Duration
	emptyValueExpireDuration time.Duration
	bit1                     int64
}

func (o Options) asyncFailBack() bool {
	return o.bit1&asyncFailBack == asyncFailBack
}

func (o Options) mustToBaser() bool {
	return o.bit1&mustToBase == mustToBase
}

func (o Options) isSetCache() bool {
	return o.bit1&setCache == setCache
}

func (o Options) syncSetCache() bool {
	return o.bit1&syncSetCache == syncSetCache
}

func (o Options) isSetEmptyValueCache() bool {
	return o.bit1&setEmptyValueCache == setEmptyValueCache
}

func (o Options) expireTime() time.Duration {
	return o.expireDuration
}

func (o Options) emptyValueExpireTime() time.Duration {
	expireTime := o.emptyValueExpireDuration
	if expireTime <= 0 {
		expireTime = time.Minute * 20
	}
	return expireTime
}
