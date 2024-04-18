package handler

import (
	"context"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/internal/cache"
	"github.com/grassrootseconomics/celo-tracker/internal/emitter"
	"github.com/grassrootseconomics/w3-celo"
)

type (
	IndexRemoveHandler struct {
		topicHash common.Hash
		event     *w3.Event
		cache     cache.Cache
	}
)

var (
	indexRemoveTopicHash = w3.H("0x24a12366c02e13fe4a9e03d86a8952e85bb74a456c16e4a18b6d8295700b74bb")
	indexRemoveEvent     = w3.MustNewEvent("AddressRemoved(address _token)")
	indexRemoveSig       = w3.MustNewFunc("remove(address)", "bool")
)

func (h *IndexRemoveHandler) HandleLog(ctx context.Context, msg LogMessage, emitter emitter.Emitter) error {
	if msg.Log.Topics[0] == indexRemoveTopicHash {
		var (
			address common.Address
		)

		if err := indexRemoveEvent.DecodeArgs(msg.Log, &address); err != nil {
			return err
		}

		indexRemoveEvent := Event{
			Block:           msg.Log.BlockNumber,
			ContractAddress: msg.Log.Address.Hex(),
			Success:         true,
			Timestamp:       msg.BlockTime,
			TxHash:          msg.Log.TxHash.Hex(),
			TxType:          "INDEX_REMOVE",
			Payload: map[string]any{
				"address": address.Hex(),
			},
		}

		if h.cache.ISWatchAbleIndex(address.Hex()) {
			h.cache.Remove(address.Hex())
		}

		return emitter.Emit(ctx, indexRemoveEvent)
	}

	return nil
}

func (h *IndexRemoveHandler) HandleRevert(ctx context.Context, msg RevertMessage, emitter emitter.Emitter) error {
	if len(msg.InputData) < 8 {
		return nil
	}

	switch msg.InputData[:8] {
	case "29092d0e":
		var (
			address common.Address
		)

		if err := indexRemoveSig.DecodeArgs(w3.B(msg.InputData), &address); err != nil {
			return err
		}

		indexRemoveEvent := Event{
			Block:           msg.Block,
			ContractAddress: msg.ContractAddress,
			Success:         false,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.TxHash,
			TxType:          "INDEX_REMOVE",
			Payload: map[string]any{
				"revertReason": msg.RevertReason,
				"address":      address.Hex(),
			},
		}

		return emitter.Emit(ctx, indexRemoveEvent)
	}

	return nil
}
