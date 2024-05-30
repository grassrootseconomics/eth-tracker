package pub

import (
	"context"
	"log/slog"

	"github.com/grassrootseconomics/celo-tracker/pkg/event"
)

type consolePub struct {
	logg *slog.Logger
}

func NewConsolePub(logg *slog.Logger) Pub {
	return &consolePub{
		logg: logg,
	}
}

func (p *consolePub) Send(_ context.Context, payload event.Event) error {
	data, err := payload.Serialize()
	if err != nil {
		return err
	}

	p.logg.Info("emitted event", "json_payload", string(data))
	return nil
}

func (p *consolePub) Close() {}
