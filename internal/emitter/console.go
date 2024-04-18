package emitter

import (
	"context"
	"encoding/json"
	"log/slog"
)

type (
	ConsoleEmitter struct {
		logg *slog.Logger
	}
)

func NewConsoleEmitter(logg *slog.Logger) *ConsoleEmitter {
	return &ConsoleEmitter{
		logg: logg,
	}
}

func (l *ConsoleEmitter) Emit(_ context.Context, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	l.logg.Info("emitted event", "json_payload", string(jsonData))
	return nil
}
