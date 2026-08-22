package kv

import "github.com/puzpuzpuz/xsync/v4"

type xmap struct {
	m *xsync.Map[string, KVVal]
}

var _ KV = &xmap{}

func newXmap() *xmap {
	return &xmap{
		m: xsync.NewMap[string, KVVal](),
	}
}

func (kv *xmap) Set(p KVSetParam) error {
	kv.m.Store(p.key.raw, p.val)
	return nil
}

func (kv *xmap) Get(p KVGetParam) (val KVVal, err error) {
	kv.m.Compute(
		p.key.raw,
		func(current KVVal, loaded bool) (after KVVal, op xsync.ComputeOp) {
			val = current
			switch {
			case !loaded:
				err = ErrKVNotFound
				return current, xsync.CancelOp
			case current.Expired():
				err = ERrKVExpired
				return current, xsync.DeleteOp
			default:
				return current, xsync.CancelOp
			}
		})

	return
}

func (kv *xmap) Del(p KVDelParam) error {
	kv.m.Delete(p.key.raw)
	return nil
}
