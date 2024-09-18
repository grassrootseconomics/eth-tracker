package router

import (
	"context"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
)

type (
	Callback    func(context.Context, event.Event) error
	HandlerFunc func(context.Context, interface{}, Callback) error

	LogPayload struct {
		Log       *types.Log
		Timestamp uint64
	}

	LogRouteEntry struct {
		Name        string
		Signature   common.Hash
		HandlerFunc HandlerFunc
	}

	InputDataPayload struct {
		From            string
		InputData       string
		Block           uint64
		ContractAddress string
		Timestamp       uint64
		TxHash          string
	}

	InputDataEntry struct {
		Name        string
		Signature   string
		HandlerFunc HandlerFunc
	}

	Router struct {
		logHandlers       map[common.Hash]LogRouteEntry
		inputDataHandlers map[string]InputDataEntry
	}
)

func New() *Router {
	return &Router{
		logHandlers:       make(map[common.Hash]LogRouteEntry),
		inputDataHandlers: make(map[string]InputDataEntry),
	}
}

func (r *Router) RegisterLogRoute(signature common.Hash, handlerFunc HandlerFunc) {
	r.logHandlers[signature] = LogRouteEntry{
		Signature:   signature,
		HandlerFunc: handlerFunc,
	}
}

func (r *Router) RegisterInputDataRoute(signature string, handlerFunc HandlerFunc) {
	r.inputDataHandlers[signature] = InputDataEntry{
		Signature:   signature,
		HandlerFunc: handlerFunc,
	}
}

func (r *Router) ProcessLog(ctx context.Context, payload LogPayload, cb Callback) error {
	handler, ok := r.logHandlers[payload.Log.Topics[0]]
	if ok {
		return handler.HandlerFunc(ctx, payload, cb)
	}

	return nil
}

func (r *Router) ProcessInputData(ctx context.Context, payload InputDataPayload, cb Callback) error {
	if len(payload.InputData) < 8 {
		return nil
	}

	handler, ok := r.inputDataHandlers[payload.InputData[:8]]
	if ok {
		return handler.HandlerFunc(ctx, payload, cb)
	}

	return nil
}
