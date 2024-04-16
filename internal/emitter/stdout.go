package emitter

import (
	"context"
	"encoding/json"
	"log/slog"
)

type (
	LogEmitter struct {
		logg *slog.Logger
	}
)

func (l *LogEmitter) Emit(_ context.Context, payload []byte) error {
	var event map[string]interface{}

	if err := json.Unmarshal(payload, &event); err != nil {
		return err
	}

	l.logg.Info("emitted event", "json_payload", event)
	return nil
}
