package pool

import (
	"github.com/gammazero/workerpool"
)

type (
	PoolOpts struct {
		PoolSize int
	}
)

func NewPool(o PoolOpts) *workerpool.WorkerPool {
	return workerpool.New(o.PoolSize)
}
