package handler

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/internal/event"
	"github.com/grassrootseconomics/celo-tracker/internal/pub"
	"github.com/grassrootseconomics/w3-celo"
)

type (
	TokenTransferHandler struct {
		topicHash common.Hash
		event     *w3.Event
	}
)

const (
	transferEventName = "TOKEN_TRANSFER"
)

var (
	tokenTransferTopicHash = w3.H("0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef")
	tokenTransferEvent     = w3.MustNewEvent("Transfer(address indexed _from, address indexed _to, uint256 _value)")
	tokenTransferSig       = w3.MustNewFunc("transfer(address, uint256)", "bool")
	tokenTransferFromSig   = w3.MustNewFunc("transferFrom(address, address, uint256)", "bool")
)

func (h *TokenTransferHandler) Name() string {
	return transferEventName
}

func (h *TokenTransferHandler) HandleLog(ctx context.Context, msg LogMessage, pub pub.Pub) error {
	if msg.Log.Topics[0] == tokenTransferTopicHash {
		var (
			from  common.Address
			to    common.Address
			value big.Int
		)

		if err := tokenTransferEvent.DecodeArgs(msg.Log, &from, &to, &value); err != nil {
			return err
		}

		tokenTransferEvent := event.Event{
			Block:           msg.Log.BlockNumber,
			ContractAddress: msg.Log.Address.Hex(),
			Success:         true,
			Timestamp:       msg.BlockTime,
			TxHash:          msg.Log.TxHash.Hex(),
			TxType:          transferEventName,
			Payload: map[string]any{
				"from":  from.Hex(),
				"to":    to.Hex(),
				"value": value.String(),
			},
		}

		return pub.Send(ctx, tokenTransferEvent)
	}

	return nil
}

func (h *TokenTransferHandler) HandleRevert(ctx context.Context, msg RevertMessage, pub pub.Pub) error {
	if len(msg.InputData) < 8 {
		return nil
	}

	switch msg.InputData[:8] {
	case "a9059cbb":
		var (
			to    common.Address
			value big.Int
		)

		if err := tokenTransferSig.DecodeArgs(w3.B(msg.InputData), &to, &value); err != nil {
			return err
		}

		tokenTransferEvent := event.Event{
			Block:           msg.Block,
			ContractAddress: msg.ContractAddress,
			Success:         false,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.TxHash,
			TxType:          transferEventName,
			Payload: map[string]any{
				"revertReason": msg.RevertReason,
				"from":         msg.From,
				"to":           to.Hex(),
				"value":        value.String(),
			},
		}

		return pub.Send(ctx, tokenTransferEvent)
	case "23b872dd":
		var (
			from  common.Address
			to    common.Address
			value big.Int
		)

		if err := tokenTransferFromSig.DecodeArgs(w3.B(msg.InputData), &from, &to, &value); err != nil {
			return err
		}

		tokenTransferEvent := event.Event{
			Block:           msg.Block,
			ContractAddress: msg.ContractAddress,
			Success:         false,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.TxHash,
			TxType:          transferEventName,
			Payload: map[string]any{
				"revertReason": msg.RevertReason,
				"from":         from.Hex(),
				"to":           to.Hex(),
				"value":        value.String(),
			},
		}

		return pub.Send(ctx, tokenTransferEvent)
	}

	return nil
}
