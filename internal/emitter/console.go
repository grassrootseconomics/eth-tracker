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

func (l *ConsoleEmitter) Emit(_ context.Context, payload []byte) error {
	var event map[string]interface{}

	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}

	l.logg.Info("emitted event", "json_payload", event)
	return nil
}
