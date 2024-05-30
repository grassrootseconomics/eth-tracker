package handler

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/internal/pub"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/w3-celo"
)

type tokenMintHandler struct {
	pub pub.Pub
}

const mintEventName = "TOKEN_MINT"

var (
	tokenMintTopicHash = w3.H("0xab8530f87dc9b59234c4623bf917212bb2536d647574c8e7e5da92c2ede0c9f8")
	tokenMintEvent     = w3.MustNewEvent("Mint(address indexed _tokenMinter, address indexed _beneficiary, uint256 _value)")
	tokenMintToSig     = w3.MustNewFunc("MintTo(address, uint256)", "bool")
)

func NewTokenMintHandler(pub pub.Pub) *tokenMintHandler {
	return &tokenMintHandler{
		pub: pub,
	}
}

func (h *tokenMintHandler) Name() string {
	return mintEventName
}

func (h *tokenMintHandler) HandleLog(ctx context.Context, msg LogMessage) error {
	if msg.Log.Topics[0] == tokenMintTopicHash {
		var (
			tokenMinter common.Address
			to          common.Address
			value       big.Int
		)

		if err := tokenMintEvent.DecodeArgs(msg.Log, &tokenMinter, &to, &value); err != nil {
			return err
		}

		tokenMintEvent := event.Event{
			Index:           msg.Log.Index,
			Block:           msg.Log.BlockNumber,
			ContractAddress: msg.Log.Address.Hex(),
			Success:         true,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.Log.TxHash.Hex(),
			TxType:          mintEventName,
			Payload: map[string]any{
				"tokenMinter": tokenMinter.Hex(),
				"to":          to.Hex(),
				"value":       value.String(),
			},
		}

		return h.pub.Send(ctx, tokenMintEvent)
	}

	return nil
}

func (h *tokenMintHandler) HandleRevert(ctx context.Context, msg RevertMessage) error {
	if len(msg.InputData) < 8 {
		return nil
	}

	switch msg.InputData[:8] {
	case "449a52f8":
		var (
			to    common.Address
			value big.Int
		)

		if err := tokenMintToSig.DecodeArgs(w3.B(msg.InputData), &to, &value); err != nil {
			return err
		}

		tokenMintEvent := event.Event{
			Block:           msg.Block,
			ContractAddress: msg.ContractAddress,
			Success:         false,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.TxHash,
			TxType:          mintEventName,
			Payload: map[string]any{
				"revertReason": msg.RevertReason,
				"tokenMinter":  msg.From,
				"to":           to.Hex(),
				"value":        value.String(),
			},
		}

		return h.pub.Send(ctx, tokenMintEvent)
	}

	return nil
}
