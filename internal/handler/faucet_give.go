package handler

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/internal/emitter"
	"github.com/grassrootseconomics/w3-celo"
)

type (
	FaucetGiveHandler struct {
		topicHash common.Hash
		event     *w3.Event
	}
)

var (
	faucetGiveTopicHash = w3.H("0x26162814817e23ec5035d6a2edc6c422da2da2119e27cfca6be65cc2dc55ca4c")
	faucetGiveEvent     = w3.MustNewEvent("Give(address indexed _recipient, address indexed _token, uint256 _amount)")
	faucetGiveToSig     = w3.MustNewFunc("giveTo(address)", "uint256")
	faucetGimmeSig      = w3.MustNewFunc("gimme()", "uint256")
)

func (h *FaucetGiveHandler) HandleLog(ctx context.Context, msg LogMessage, emitter emitter.Emitter) error {
	if msg.Log.Topics[0] == faucetGiveTopicHash {
		var (
			recipient common.Address
			token     common.Address
			amount    big.Int
		)

		if err := faucetGiveEvent.DecodeArgs(msg.Log, &recipient, &token, &amount); err != nil {
			return err
		}

		faucetGiveEvent := Event{
			Block:           msg.Log.BlockNumber,
			ContractAddress: msg.Log.Address.Hex(),
			Success:         true,
			Timestamp:       msg.BlockTime,
			TxHash:          msg.Log.TxHash.Hex(),
			TxType:          "FAUCET_GIVE",
			Payload: map[string]any{
				"recipient": recipient.Hex(),
				"token":     token.Hex(),
				"amount":    amount.String(),
			},
		}

		return emitter.Emit(ctx, faucetGiveEvent)
	}

	return nil
}

func (h *FaucetGiveHandler) HandleRevert(ctx context.Context, msg RevertMessage, emitter emitter.Emitter) error {
	if len(msg.InputData) < 8 {
		return nil
	}

	switch msg.InputData[:8] {
	case "63e4bff4":
		var (
			to common.Address
		)

		if err := faucetGiveToSig.DecodeArgs(w3.B(msg.InputData), &to); err != nil {
			return err
		}

		faucetGiveEvent := Event{
			Block:           msg.Block,
			ContractAddress: msg.ContractAddress,
			Success:         false,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.TxHash,
			TxType:          "FAUCET_GIVE",
			Payload: map[string]any{
				"revertReason": msg.RevertReason,
				"recipient":    to.Hex(),
				"token":        common.ZeroAddress.Hex(),
				"amount":       "0",
			},
		}

		return emitter.Emit(ctx, faucetGiveEvent)
	case "de82efb4":
		faucetGiveEvent := Event{
			Block:           msg.Block,
			ContractAddress: msg.ContractAddress,
			Success:         false,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.TxHash,
			TxType:          "FAUCET_GIVE",
			Payload: map[string]any{
				"revertReason": msg.RevertReason,
				"recipient":    common.ZeroAddress.Hex(),
				"token":        common.ZeroAddress.Hex(),
				"amount":       "0",
			},
		}

		return emitter.Emit(ctx, faucetGiveEvent)
	}

	return nil
}
