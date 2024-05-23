package handler

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/internal/event"
	"github.com/grassrootseconomics/celo-tracker/internal/pub"
	"github.com/grassrootseconomics/w3-celo"
)

type tokenBurnHandler struct {
	pub pub.Pub
}

const burnEventName = "TOKEN_BURN"

var (
	tokenBurnTopicHash = w3.H("0xcc16f5dbb4873280815c1ee09dbd06736cffcc184412cf7a71a0fdb75d397ca5")
	tokenBurnEvent     = w3.MustNewEvent("tokenBurn(address indexed _tokenBurner, uint256 _value)")
	tokenBurnToSig     = w3.MustNewFunc("tokenBurn(uint256)", "bool")
)

func NewTokenBurnHandler(pub pub.Pub) *tokenBurnHandler {
	return &tokenBurnHandler{
		pub: pub,
	}
}

func (h *tokenBurnHandler) Name() string {
	return burnEventName
}

func (h *tokenBurnHandler) HandleLog(ctx context.Context, msg LogMessage) error {
	if msg.Log.Topics[0] == tokenBurnTopicHash {
		var (
			tokenBurner common.Address
			value       big.Int
		)

		if err := tokenBurnEvent.DecodeArgs(msg.Log, &tokenBurner, &value); err != nil {
			return err
		}

		tokenBurnEvent := event.Event{
			Index:           msg.Log.Index,
			Block:           msg.Log.BlockNumber,
			ContractAddress: msg.Log.Address.Hex(),
			Success:         true,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.Log.TxHash.Hex(),
			TxType:          burnEventName,
			Payload: map[string]any{
				"tokenBurner": tokenBurner.Hex(),
				"value":       value.String(),
			},
		}

		return h.pub.Send(ctx, tokenBurnEvent)
	}

	return nil
}

func (h *tokenBurnHandler) HandleRevert(ctx context.Context, msg RevertMessage) error {
	if len(msg.InputData) < 8 {
		return nil
	}

	switch msg.InputData[:8] {
	case "42966c68":
		var value big.Int

		if err := tokenBurnToSig.DecodeArgs(w3.B(msg.InputData), &value); err != nil {
			return err
		}

		tokenBurnEvent := event.Event{
			Block:           msg.Block,
			ContractAddress: msg.ContractAddress,
			Success:         false,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.TxHash,
			TxType:          burnEventName,
			Payload: map[string]any{
				"revertReason": msg.RevertReason,
				"tokenBurner":  msg.From,
				"value":        value.String(),
			},
		}

		return h.pub.Send(ctx, tokenBurnEvent)
	}

	return nil
}
