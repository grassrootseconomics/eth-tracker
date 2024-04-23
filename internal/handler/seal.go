package handler

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/internal/pub"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/w3-celo"
)

type (
	SealHandler struct {
		topicHash common.Hash
		event     *w3.Event
	}
)

const (
	sealEventName = "SEAL_STATE_CHANGE"
)

var (
	sealTopicHash = w3.H("0x6b7e2e653f93b645d4ed7292d6429f96637084363e477c8aaea1a43ed13c284e")
	sealEvent     = w3.MustNewEvent("SealStateChange(bool indexed _final, uint256 _sealState)")
	sealToSig     = w3.MustNewFunc("seal(uint256)", "uint256")
)

func (h *SealHandler) Name() string {
	return sealEventName
}

func (h *SealHandler) HandleLog(ctx context.Context, msg LogMessage, pub pub.Pub) error {
	if msg.Log.Topics[0] == sealTopicHash {
		var (
			final     bool
			sealState big.Int
		)

		if err := sealEvent.DecodeArgs(msg.Log, &final, &sealState); err != nil {
			return err
		}

		sealEvent := event.Event{
			Block:           msg.Log.BlockNumber,
			ContractAddress: msg.Log.Address.Hex(),
			Success:         true,
			Timestamp:       msg.BlockTime,
			TxHash:          msg.Log.TxHash.Hex(),
			TxType:          sealEventName,
			Payload: map[string]any{
				"final":     final,
				"sealState": sealState.String(),
			},
		}

		return pub.Send(ctx, sealEvent)
	}

	return nil
}

func (h *SealHandler) HandleRevert(ctx context.Context, msg RevertMessage, pub pub.Pub) error {
	if len(msg.InputData) < 8 {
		return nil
	}

	switch msg.InputData[:8] {
	case "86fe212d":
		var (
			sealState big.Int
		)

		if err := sealToSig.DecodeArgs(w3.B(msg.InputData), &sealState); err != nil {
			return err
		}

		sealEvent := event.Event{
			Block:           msg.Block,
			ContractAddress: msg.ContractAddress,
			Success:         false,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.TxHash,
			TxType:          sealEventName,
			Payload: map[string]any{
				"revertReason": msg.RevertReason,
				"sealState":    sealState.String(),
			},
		}

		return pub.Send(ctx, sealEvent)
	}

	return nil
}
