package handler

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/internal/event"
	"github.com/grassrootseconomics/celo-tracker/internal/pub"
	"github.com/grassrootseconomics/w3-celo"
)

type poolDepositHandler struct {
	pub pub.Pub
}

const poolDepositEventName = "POOL_DEPOSIT"

var (
	poolDepositTopicHash = w3.H("0x5548c837ab068cf56a2c2479df0882a4922fd203edb7517321831d95078c5f62")
	poolDepositEvent     = w3.MustNewEvent("Deposit(address indexed initiator, address indexed tokenIn, uint256 amountIn)")
	poolDepositSig       = w3.MustNewFunc("deposit(address, uint256)", "")
)

func NewPoolDepositHandler(pub pub.Pub) *poolDepositHandler {
	return &poolDepositHandler{
		pub: pub,
	}
}

func (h *poolDepositHandler) Name() string {
	return poolDepositEventName
}

func (h *poolDepositHandler) HandleLog(ctx context.Context, msg LogMessage) error {
	if msg.Log.Topics[0] == poolDepositTopicHash {
		var (
			initiator common.Address
			tokenIn   common.Address
			amountIn  big.Int
		)

		if err := poolDepositEvent.DecodeArgs(
			msg.Log,
			&initiator,
			&tokenIn,
			&amountIn,
		); err != nil {
			return err
		}

		poolDepositEvent := event.Event{
			Index:           msg.Log.Index,
			Block:           msg.Log.BlockNumber,
			ContractAddress: msg.Log.Address.Hex(),
			Success:         true,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.Log.TxHash.Hex(),
			TxType:          poolDepositEventName,
			Payload: map[string]any{
				"initiator": initiator.Hex(),
				"tokenIn":   tokenIn.Hex(),
				"amountIn":  amountIn.String(),
			},
		}

		return h.pub.Send(ctx, poolDepositEvent)
	}

	return nil
}

func (h *poolDepositHandler) HandleRevert(ctx context.Context, msg RevertMessage) error {
	if len(msg.InputData) < 8 {
		return nil
	}

	switch msg.InputData[:8] {
	case "47e7ef24":
		var (
			tokenIn  common.Address
			amountIn big.Int
		)

		if err := poolDepositSig.DecodeArgs(w3.B(msg.InputData), &tokenIn, &amountIn); err != nil {
			return err
		}

		poolDepositEvent := event.Event{
			Block:           msg.Block,
			ContractAddress: msg.ContractAddress,
			Success:         false,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.TxHash,
			TxType:          poolDepositEventName,
			Payload: map[string]any{
				"revertReason": msg.RevertReason,
				"initiator":    msg.From,
				"tokenIn":      tokenIn.Hex(),
				"amountIn":     amountIn.String(),
			},
		}

		return h.pub.Send(ctx, poolDepositEvent)
	}

	return nil
}
