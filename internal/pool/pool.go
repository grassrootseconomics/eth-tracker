package pool

import (
	"log/slog"
	"runtime/debug"

	"github.com/alitto/pond"
)

type PoolOpts struct {
	Logg        *slog.Logger
	WorkerCount int
}

func NewPool(o PoolOpts) *pond.WorkerPool {
	return pond.New(
		o.WorkerCount,
		1,
		pond.Strategy(pond.Balanced()),
		pond.PanicHandler(panicHandler(o.Logg)),
	)
}

func panicHandler(logg *slog.Logger) func(interface{}) {
	return func(panic interface{}) {
		logg.Error("block processor goroutine exited from a panic", "error", panic, "stack_trace", string(debug.Stack()))
	}
}
