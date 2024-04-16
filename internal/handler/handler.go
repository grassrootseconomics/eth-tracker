package handler

import (
	"context"

	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/grassrootseconomics/w3-celo"
)

type (
	EmitterEmitFunc func(context.Context, []byte) error

	Handler interface {
		Handle(context.Context, *types.Log, EmitterEmitFunc) error
	}
)

func New() []Handler {
	transferHandler := &TransferHandler{
		topicHash: w3.H("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"),
		event:     w3.MustNewEvent("Transfer(address indexed _from, address indexed _to, uint256 _value)"),
	}

	return []Handler{
		transferHandler,
	}
}
