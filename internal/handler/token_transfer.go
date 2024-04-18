package handler

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/internal/emitter"
	"github.com/grassrootseconomics/w3-celo"
)

type (
	TransferHandler struct {
		topicHash common.Hash
		event     *w3.Event
	}
)

var (
	transferTopicHash = w3.H("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	transferEvent     = w3.MustNewEvent("Transfer(address indexed _from, address indexed _to, uint256 _value)")
	transferSig       = w3.MustNewFunc("transfer(address, uint256)", "bool")
	transferFromSig   = w3.MustNewFunc("transferFrom(address, address, uint256)", "bool")
)

func (h *TransferHandler) HandleLog(ctx context.Context, msg LogMessage, emitter emitter.Emitter) error {
	if msg.Log.Topics[0] == transferTopicHash {
		var (
			from  common.Address
			to    common.Address
			value big.Int
		)

		if err := transferEvent.DecodeArgs(msg.Log, &from, &to, &value); err != nil {
			return err
		}

		transferEvent := Event{
			Block:           msg.Log.BlockNumber,
			ContractAddress: msg.Log.Address.Hex(),
			Success:         true,
			Timestamp:       msg.BlockTime,
			TxHash:          msg.Log.TxHash.Hex(),
			TxType:          "TRANSFER",
			Payload: map[string]any{
				"from":  from.Hex(),
				"to":    to.Hex(),
				"value": value.String(),
			},
		}

		return emitter.Emit(ctx, transferEvent)
	}

	return nil
}

func (h *TransferHandler) HandleRevert(ctx context.Context, msg RevertMessage, emitter emitter.Emitter) error {
	if len(msg.InputData) < 8 {
		return nil
	}

	switch msg.InputData[:8] {
	case "a9059cbb":
		var (
			to    common.Address
			value big.Int
		)

		if err := transferSig.DecodeArgs(w3.B(msg.InputData), &to, &value); err != nil {
			return err
		}

		transferEvent := Event{
			Block:           msg.Block,
			ContractAddress: msg.ContractAddress,
			Success:         false,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.TxHash,
			TxType:          "TRANSFER",
			Payload: map[string]any{
				"revertReason": msg.RevertReason,
				"from":         msg.From,
				"to":           to.Hex(),
				"value":        value.String(),
			},
		}

		return emitter.Emit(ctx, transferEvent)
	case "23b872dd":
		var (
			from  common.Address
			to    common.Address
			value big.Int
		)

		if err := transferFromSig.DecodeArgs(w3.B(msg.InputData), &from, &to, &value); err != nil {
			return err
		}

		transferEvent := Event{
			Block:           msg.Block,
			ContractAddress: msg.ContractAddress,
			Success:         false,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.TxHash,
			TxType:          "TRANSFER",
			Payload: map[string]any{
				"revertReason": msg.RevertReason,
				"from":         from.Hex(),
				"to":           to.Hex(),
				"value":        value.String(),
			},
		}

		return emitter.Emit(ctx, transferEvent)
	}

	return nil
}
