package handler

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/internal/emitter"
	"github.com/grassrootseconomics/w3-celo"
)

type (
	BurnHandler struct {
		topicHash common.Hash
		event     *w3.Event
	}
)

var (
	burnTopicHash = w3.H("0xcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5")
	burnEvent     = w3.MustNewEvent("Burn(address indexed _burner, uint256 _value)")
	burnToSig     = w3.MustNewFunc("burn(uint256)", "bool")
)

func (h *BurnHandler) HandleLog(ctx context.Context, msg LogMessage, emitter emitter.Emitter) error {
	if msg.Log.Topics[0] == burnTopicHash {
		var (
			burner common.Address
			value  big.Int
		)

		if err := burnEvent.DecodeArgs(msg.Log, &burner, &value); err != nil {
			return err
		}

		burnEvent := Event{
			Block:           msg.Log.BlockNumber,
			ContractAddress: msg.Log.Address.Hex(),
			Success:         true,
			Timestamp:       msg.BlockTime,
			TxHash:          msg.Log.TxHash.Hex(),
			TxType:          "BURN",
			Payload: map[string]any{
				"burner": burner.Hex(),
				"value":  value.String(),
			},
		}

		return emitter.Emit(ctx, burnEvent)
	}

	return nil
}

func (h *BurnHandler) HandleRevert(ctx context.Context, msg RevertMessage, emitter emitter.Emitter) error {
	if len(msg.InputData) < 8 {
		return nil
	}

	switch msg.InputData[:8] {
	case "42966c68":
		var (
			value big.Int
		)

		if err := burnToSig.DecodeArgs(w3.B(msg.InputData), &value); err != nil {
			return err
		}

		burnEvent := Event{
			Block:           msg.Block,
			ContractAddress: msg.ContractAddress,
			Success:         false,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.TxHash,
			TxType:          "BURN",
			Payload: map[string]any{
				"revertReason": msg.RevertReason,
				"burner":       msg.From,
				"value":        value.String(),
			},
		}

		return emitter.Emit(ctx, burnEvent)
	}

	return nil
}
