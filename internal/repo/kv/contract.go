package kv

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrKVNotFound = errors.New("repo: kv: not found")
	errKVExpired  = fmt.Errorf("%w: expired", ErrKVNotFound)
)

type KV interface {
	Set(KVSetParam) error
	Get(KVGetParam) (KVVal, error)
	Del(KVDelParam) error
}

func NewKV() KV {
	return newXmap()
}

type KVSetParam struct {
	key KVKey
	val KVVal
}

func NewKVSetParam(key KVKey, val KVVal) KVSetParam {
	return KVSetParam{key: key, val: val}
}

type KVGetParam struct {
	key KVKey
}

func NewKVGetParam(key KVKey) KVGetParam {
	return KVGetParam{key: key}
}

type KVDelParam struct {
	key KVKey
}

func NewKVDelParam(key KVKey) KVDelParam {
	return KVDelParam{key: key}
}

type KVKey struct {
	raw string
}

func NewKVKey(raw fmt.Stringer) KVKey {
	return KVKey{raw: raw.String()}
}

type KVVal struct {
	raw string

	exp *time.Time
	lru time.Time
}

func NewKVVal(raw fmt.Stringer, exp *time.Time) KVVal {
	return KVVal{raw: raw.String(), exp: exp, lru: time.Now()}
}

func (v KVVal) Expired() bool {
	return v.expiredAt(time.Now())
}

func (v KVVal) expiredAt(now time.Time) bool {
	return v.exp != nil && !now.Before(*v.exp)
}

func (v KVVal) String() string {
	return v.raw
}
