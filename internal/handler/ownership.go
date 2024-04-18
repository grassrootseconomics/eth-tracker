package handler

import (
	"context"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/internal/emitter"
	"github.com/grassrootseconomics/w3-celo"
)

type (
	OwnershipHandler struct {
		topicHash common.Hash
		event     *w3.Event
	}
)

var (
	ownershipTopicHash = w3.H("0x8be0079c531659141344cd1fd0a4f28419497f9722a3daafe3b4186f6b6457e0")
	ownershipEvent     = w3.MustNewEvent("OwnershipTransferred(address indexed previousOwner, address indexed newOwner)")
	ownershipToSig     = w3.MustNewFunc("transferOwnership(address)", "bool")
)

func (h *OwnershipHandler) HandleLog(ctx context.Context, msg LogMessage, emitter emitter.Emitter) error {
	if msg.Log.Topics[0] == ownershipTopicHash {
		var (
			previousOwner common.Address
			newOwner      common.Address
		)

		if err := ownershipEvent.DecodeArgs(msg.Log, &previousOwner, &newOwner); err != nil {
			return err
		}

		ownershipEvent := Event{
			Block:           msg.Log.BlockNumber,
			ContractAddress: msg.Log.Address.Hex(),
			Success:         true,
			Timestamp:       msg.BlockTime,
			TxHash:          msg.Log.TxHash.Hex(),
			TxType:          "OWNERSHIP_TRANSFERRED",
			Payload: map[string]any{
				"previousOwner": previousOwner.Hex(),
				"newOwner":      newOwner.Hex(),
			},
		}

		return emitter.Emit(ctx, ownershipEvent)
	}

	return nil
}

func (h *OwnershipHandler) HandleRevert(ctx context.Context, msg RevertMessage, emitter emitter.Emitter) error {
	if len(msg.InputData) < 8 {
		return nil
	}

	switch msg.InputData[:8] {
	case "f2fde38b":
		var (
			newOwner common.Address
		)

		if err := ownershipToSig.DecodeArgs(w3.B(msg.InputData), &newOwner); err != nil {
			return err
		}

		ownershipEvent := Event{
			Block:           msg.Block,
			ContractAddress: msg.ContractAddress,
			Success:         false,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.TxHash,
			TxType:          "ownership",
			Payload: map[string]any{
				"revertReason":  msg.RevertReason,
				"previousOwner": msg.From,
				"newOwner":      newOwner.Hex(),
			},
		}

		return emitter.Emit(ctx, ownershipEvent)
	}

	return nil
}
