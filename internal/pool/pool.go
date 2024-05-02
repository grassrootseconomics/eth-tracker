package pool

import (
	"runtime"

	"github.com/gammazero/workerpool"
)

func NewPool() *workerpool.WorkerPool {
	return workerpool.New(runtime.NumCPU())
}
