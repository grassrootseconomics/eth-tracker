package pool

import (
	"log/slog"
	"runtime"
	"runtime/debug"

	"github.com/alitto/pond"
)

const (
	nProcFactor = 5
)

func NewPool(logg *slog.Logger) *pond.WorkerPool {
	return pond.New(
		runtime.NumCPU()*nProcFactor,
		1,
		pond.Strategy(pond.Balanced()),
		pond.PanicHandler(panicHandler(logg)),
	)
}

func panicHandler(logg *slog.Logger) func(interface{}) {
	return func(panic interface{}) {
		logg.Error("block processor worker exited from a panic", "error", panic, "stack_trace", string(debug.Stack()))
	}
}
