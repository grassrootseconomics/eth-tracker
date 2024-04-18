package handler

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/internal/emitter"
	"github.com/grassrootseconomics/w3-celo"
)

type (
	MintHandler struct {
		topicHash common.Hash
		event     *w3.Event
	}
)

var (
	mintTopicHash = w3.H("0xab8530f87dc9b59234c4623bf917212bb2536d647574c8e7e5da92c2ede0c9f8")
	mintEvent     = w3.MustNewEvent("Mint(address indexed _minter, address indexed _beneficiary, uint256 _value)")
	mintToSig     = w3.MustNewFunc("mintTo(address, uint256)", "bool")
)

func (h *MintHandler) HandleLog(ctx context.Context, msg LogMessage, emitter emitter.Emitter) error {
	if msg.Log.Topics[0] == mintTopicHash {
		var (
			minter common.Address
			to     common.Address
			value  big.Int
		)

		if err := mintEvent.DecodeArgs(msg.Log, &minter, &to, &value); err != nil {
			return err
		}

		mintEvent := Event{
			Block:           msg.Log.BlockNumber,
			ContractAddress: msg.Log.Address.Hex(),
			Success:         true,
			Timestamp:       msg.BlockTime,
			TxHash:          msg.Log.TxHash.Hex(),
			TxType:          "MINT",
			Payload: map[string]any{
				"minter": minter.Hex(),
				"to":     to.Hex(),
				"value":  value.String(),
			},
		}

		return emitter.Emit(ctx, mintEvent)
	}

	return nil
}

func (h *MintHandler) HandleRevert(ctx context.Context, msg RevertMessage, emitter emitter.Emitter) error {
	if len(msg.InputData) < 8 {
		return nil
	}

	switch msg.InputData[:8] {
	case "449a52f8":
		var (
			to    common.Address
			value big.Int
		)

		if err := mintToSig.DecodeArgs(w3.B(msg.InputData), &to, &value); err != nil {
			return err
		}

		mintEvent := Event{
			Block:           msg.Block,
			ContractAddress: msg.ContractAddress,
			Success:         false,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.TxHash,
			TxType:          "MINT",
			Payload: map[string]any{
				"revertReason": msg.RevertReason,
				"minter":       msg.From,
				"to":           to.Hex(),
				"value":        value.String(),
			},
		}

		return emitter.Emit(ctx, mintEvent)
	}

	return nil
}
