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
	PoolSwapHandler struct {
		topicHash common.Hash
		event     *w3.Event
	}
)

const (
	poolSwapEventName = "POOL_SWAP"
)

var (
	poolSwapTopicHash = w3.H("0xd6d34547c69c5ee3d2667625c188acf1006abb93e0ee7cf03925c67cf7760413")
	poolSwapEvent     = w3.MustNewEvent("Swap(address indexed initiator, address indexed tokenIn, address tokenOut, uint256 amountIn, uint256 amountOut, uint256 fee)")
	poolSwapSig       = w3.MustNewFunc("withdraw(address, address, uint256)", "")
)

func (h *PoolSwapHandler) Name() string {
	return poolSwapEventName
}

func (h *PoolSwapHandler) HandleLog(ctx context.Context, msg LogMessage, pub pub.Pub) error {
	if msg.Log.Topics[0] == poolSwapTopicHash {
		var (
			initiator common.Address
			tokenIn   common.Address
			tokenOut  common.Address
			amountIn  big.Int
			amountOut big.Int
			fee       big.Int
		)

		if err := poolSwapEvent.DecodeArgs(
			msg.Log,
			&initiator,
			&tokenIn,
			&tokenOut,
			&amountIn,
			&amountOut,
			&fee,
		); err != nil {
			return err
		}

		poolSwapEvent := event.Event{
			Block:           msg.Log.BlockNumber,
			ContractAddress: msg.Log.Address.Hex(),
			Success:         true,
			Timestamp:       msg.BlockTime,
			TxHash:          msg.Log.TxHash.Hex(),
			TxType:          poolSwapEventName,
			Payload: map[string]any{
				"initiator": initiator.Hex(),
				"tokenIn":   tokenIn.Hex(),
				"tokenOut":  tokenOut.Hex(),
				"amountIn":  amountIn.String(),
				"amountOut": amountOut.String(),
				"fee":       fee.String(),
			},
		}

		return pub.Send(ctx, poolSwapEvent)
	}

	return nil
}

func (h *PoolSwapHandler) HandleRevert(ctx context.Context, msg RevertMessage, pub pub.Pub) error {
	if len(msg.InputData) < 8 {
		return nil
	}

	switch msg.InputData[:8] {
	case "d9caed12":
		var (
			tokenOut common.Address
			tokenIn  common.Address
			amountIn big.Int
		)

		if err := poolSwapSig.DecodeArgs(w3.B(msg.InputData), &tokenOut, &tokenIn, &amountIn); err != nil {
			return err
		}

		poolSwapEvent := event.Event{
			Block:           msg.Block,
			ContractAddress: msg.ContractAddress,
			Success:         false,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.TxHash,
			TxType:          poolSwapEventName,
			Payload: map[string]any{
				"revertReason": msg.RevertReason,
				"initiator":    msg.From,
				"tokenIn":      tokenIn.Hex(),
				"tokenOut":     tokenOut.Hex(),
				"amountIn":     amountIn.String(),
				"amountOut":    "0",
				"fee":          "0",
			},
		}

		return pub.Send(ctx, poolSwapEvent)
	}

	return nil
}
