package kv

import (
	"errors"
	"time"
)

var (
	ErrKVNotFound = errors.New("repo: kv: not found")
	ERrKVExpired  = errors.New("repo: kv: expired")
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

type KVGetParam struct {
	key KVKey
}

type KVDelParam struct {
	key KVKey
}

type KVKey struct {
	raw string
}

type KVVal struct {
	raw string

	exp *time.Time
	lru *time.Time
}

func (v KVVal) Expired() bool {
	return v.expiredAt(time.Now())
}

func (v KVVal) expiredAt(now time.Time) bool {
	return v.exp != nil && !now.Before(*v.exp)
}
