package emitter

import (
	"context"
	"log/slog"
)

type (
	Emitter interface {
		Emit(context.Context, []byte) error
	}

	EmitterOpts struct {
		Logg *slog.Logger
	}
)

func New(o EmitterOpts) Emitter {
	return NewConsoleEmitter(o.Logg)
}
