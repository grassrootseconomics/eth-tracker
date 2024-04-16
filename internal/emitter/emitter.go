package emitter

import (
	"log/slog"

	"github.com/grassrootseconomics/celo-tracker/internal/handler"
)

func New(logg *slog.Logger) handler.EmitterEmitFunc {
	stdOutEmitter := &LogEmitter{
		logg: logg,
	}

	return stdOutEmitter.Emit
}
