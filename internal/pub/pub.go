package pub

import (
	"context"

	"github.com/grassrootseconomics/celo-tracker/internal/event"
)

type Pub interface {
	Send(context.Context, event.Event) error
	Close()
}
