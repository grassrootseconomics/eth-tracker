package pub

import (
	"context"
	"log/slog"

	"github.com/grassrootseconomics/celo-tracker/internal/event"
)

type (
	ConsolePub struct {
		logg *slog.Logger
	}
)

func NewConsolePub(logg *slog.Logger) Pub {
	return &ConsolePub{
		logg: logg,
	}
}

func (p *ConsolePub) Send(_ context.Context, payload event.Event) error {
	data, err := payload.Serialize()
	if err != nil {
		return err
	}

	p.logg.Info("emitted event", "json_payload", string(data))
	return nil
}

func (p *ConsolePub) Close() {}
