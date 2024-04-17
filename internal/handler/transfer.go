package handler

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/grassrootseconomics/celo-tracker/internal/emitter"
	"github.com/grassrootseconomics/w3-celo"
)

type (
	TransferHandler struct {
		topicHash common.Hash
		event     *w3.Event
	}

	TransferEvent struct {
		Contract string
		From     string
		To       string
		Value    uint64
	}
)

func (h *TransferHandler) Handle(ctx context.Context, log *types.Log, emitter emitter.Emitter) error {
	if log.Topics[0] == h.topicHash {
		var (
			from  common.Address
			to    common.Address
			value big.Int
		)

		if err := h.event.DecodeArgs(log, &from, &to, &value); err != nil {
			return err
		}

		transferEvent := &TransferEvent{
			Contract: log.Address.Hex(),
			From:     from.Hex(),
			To:       to.Hex(),
			Value:    value.Uint64(),
		}

		jsonData, err := json.Marshal(transferEvent)
		if err != nil {
			return err
		}

		return emitter.Emit(ctx, jsonData)
	}

	return nil
}
