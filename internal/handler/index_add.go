package handler

import (
	"context"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/internal/cache"
	"github.com/grassrootseconomics/celo-tracker/internal/event"
	"github.com/grassrootseconomics/celo-tracker/internal/pub"
	"github.com/grassrootseconomics/w3-celo"
)

type indexAddHandler struct {
	pub   pub.Pub
	cache cache.Cache
}

const indexAddEventName = "INDEX_ADD"

var (
	indexAddTopicHash = w3.H("0xa226db3f664042183ee0281230bba26cbf7b5057e50aee7f25a175ff45ce4d7f")
	indexAddEvent     = w3.MustNewEvent("AddressAdded(address _token)")
	indexAddSig       = w3.MustNewFunc("add(address)", "bool")
	indexRegisterSig  = w3.MustNewFunc("register(address)", "bool")
)

func NewIndexAddHandler(pub pub.Pub, cache cache.Cache) *indexAddHandler {
	return &indexAddHandler{
		pub:   pub,
		cache: cache,
	}
}

func (h *indexAddHandler) Name() string {
	return indexAddEventName
}

func (h *indexAddHandler) HandleLog(ctx context.Context, msg LogMessage) error {
	if msg.Log.Topics[0] == indexAddTopicHash {
		var address common.Address

		if err := indexAddEvent.DecodeArgs(msg.Log, &address); err != nil {
			return err
		}

		indexAddEvent := event.Event{
			Index:           msg.Log.Index,
			Block:           msg.Log.BlockNumber,
			ContractAddress: msg.Log.Address.Hex(),
			Success:         true,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.Log.TxHash.Hex(),
			TxType:          indexAddEventName,
			Payload: map[string]any{
				"address": address.Hex(),
			},
		}

		if h.cache.IsWatchableIndex(address.Hex()) {
			h.cache.Add(address.Hex(), false)
		}

		return h.pub.Send(ctx, indexAddEvent)
	}

	return nil
}

func (h *indexAddHandler) HandleRevert(ctx context.Context, msg RevertMessage) error {
	if len(msg.InputData) < 8 {
		return nil
	}

	indexAddEvent := event.Event{
		Block:           msg.Block,
		ContractAddress: msg.ContractAddress,
		Success:         false,
		Timestamp:       msg.Timestamp,
		TxHash:          msg.TxHash,
		TxType:          indexAddEventName,
	}

	switch msg.InputData[:8] {
	case "0a3b0a4f":
		var address common.Address

		indexAddEvent.Payload = map[string]any{
			"revertReason": msg.RevertReason,
			"address":      address.Hex(),
		}

		if err := indexAddSig.DecodeArgs(w3.B(msg.InputData), &address); err != nil {
			return err
		}

		return h.pub.Send(ctx, indexAddEvent)
	case "4420e486":
		var address common.Address

		indexAddEvent.Payload = map[string]any{
			"revertReason": msg.RevertReason,
			"address":      address.Hex(),
		}

		if err := indexRegisterSig.DecodeArgs(w3.B(msg.InputData), &address); err != nil {
			return err
		}

		return h.pub.Send(ctx, indexAddEvent)
	}

	return nil
}
