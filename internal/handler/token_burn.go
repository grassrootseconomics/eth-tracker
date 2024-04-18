package handler

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/internal/emitter"
	"github.com/grassrootseconomics/w3-celo"
)

type (
	TokenBurnHandler struct {
		topicHash common.Hash
		event     *w3.Event
	}
)

var (
	tokenBurnTopicHash = w3.H("0xcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5")
	tokenBurnEvent     = w3.MustNewEvent("tokenBurn(address indexed _tokenBurner, uint256 _value)")
	tokenBurnToSig     = w3.MustNewFunc("tokenBurn(uint256)", "bool")
)

func (h *TokenBurnHandler) HandleLog(ctx context.Context, msg LogMessage, emitter emitter.Emitter) error {
	if msg.Log.Topics[0] == tokenBurnTopicHash {
		var (
			tokenBurner common.Address
			value       big.Int
		)

		if err := tokenBurnEvent.DecodeArgs(msg.Log, &tokenBurner, &value); err != nil {
			return err
		}

		tokenBurnEvent := Event{
			Block:           msg.Log.BlockNumber,
			ContractAddress: msg.Log.Address.Hex(),
			Success:         true,
			Timestamp:       msg.BlockTime,
			TxHash:          msg.Log.TxHash.Hex(),
			TxType:          "TOKEN_BURN",
			Payload: map[string]any{
				"tokenBurner": tokenBurner.Hex(),
				"value":       value.String(),
			},
		}

		return emitter.Emit(ctx, tokenBurnEvent)
	}

	return nil
}

func (h *TokenBurnHandler) HandleRevert(ctx context.Context, msg RevertMessage, emitter emitter.Emitter) error {
	if len(msg.InputData) < 8 {
		return nil
	}

	switch msg.InputData[:8] {
	case "42966c68":
		var (
			value big.Int
		)

		if err := tokenBurnToSig.DecodeArgs(w3.B(msg.InputData), &value); err != nil {
			return err
		}

		tokenBurnEvent := Event{
			Block:           msg.Block,
			ContractAddress: msg.ContractAddress,
			Success:         false,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.TxHash,
			TxType:          "TOKEN_BURN",
			Payload: map[string]any{
				"revertReason": msg.RevertReason,
				"tokenBurner":  msg.From,
				"value":        value.String(),
			},
		}

		return emitter.Emit(ctx, tokenBurnEvent)
	}

	return nil
}
