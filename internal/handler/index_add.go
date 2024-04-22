package handler

import (
	"context"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/internal/cache"
	"github.com/grassrootseconomics/celo-tracker/internal/emitter"
	"github.com/grassrootseconomics/w3-celo"
)

type (
	IndexAddHandler struct {
		topicHash common.Hash
		event     *w3.Event
		cache     cache.Cache
	}
)

const (
	indexAddEventName = "INDEX_ADD"
)

var (
	indexAddTopicHash = w3.H("0xa226db3f664042183ee0281230bba26cbf7b5057e50aee7f25a175ff45ce4d7f")
	indexAddEvent     = w3.MustNewEvent("AddressAdded(address _token)")
	indexAddSig       = w3.MustNewFunc("add(address)", "bool")
	indexRegisterSig  = w3.MustNewFunc("register(address)", "bool")
)

func (h *IndexAddHandler) Name() string {
	return indexAddEventName
}

func (h *IndexAddHandler) HandleLog(ctx context.Context, msg LogMessage, emitter emitter.Emitter) error {
	if msg.Log.Topics[0] == indexAddTopicHash {
		var (
			address common.Address
		)

		if err := indexAddEvent.DecodeArgs(msg.Log, &address); err != nil {
			return err
		}

		indexAddEvent := Event{
			Block:           msg.Log.BlockNumber,
			ContractAddress: msg.Log.Address.Hex(),
			Success:         true,
			Timestamp:       msg.BlockTime,
			TxHash:          msg.Log.TxHash.Hex(),
			TxType:          indexAddEventName,
			Payload: map[string]any{
				"address": address.Hex(),
			},
		}

		if h.cache.ISWatchAbleIndex(address.Hex()) {
			h.cache.Add(address.Hex())
		}

		return emitter.Emit(ctx, indexAddEvent)
	}

	return nil
}

func (h *IndexAddHandler) HandleRevert(ctx context.Context, msg RevertMessage, emitter emitter.Emitter) error {
	if len(msg.InputData) < 8 {
		return nil
	}

	switch msg.InputData[:8] {
	case "0a3b0a4f":
		var (
			address common.Address
		)

		if err := indexAddSig.DecodeArgs(w3.B(msg.InputData), &address); err != nil {
			return err
		}

		indexAddEvent := Event{
			Block:           msg.Block,
			ContractAddress: msg.ContractAddress,
			Success:         false,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.TxHash,
			TxType:          indexAddEventName,
			Payload: map[string]any{
				"revertReason": msg.RevertReason,
				"address":      address.Hex(),
			},
		}

		return emitter.Emit(ctx, indexAddEvent)
	case "4420e486":
		var (
			address common.Address
		)

		if err := indexRegisterSig.DecodeArgs(w3.B(msg.InputData), &address); err != nil {
			return err
		}

		indexAddEvent := Event{
			Block:           msg.Block,
			ContractAddress: msg.ContractAddress,
			Success:         false,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.TxHash,
			TxType:          indexAddEventName,
			Payload: map[string]any{
				"revertReason": msg.RevertReason,
				"address":      address.Hex(),
			},
		}

		return emitter.Emit(ctx, indexAddEvent)
	}

	return nil
}
