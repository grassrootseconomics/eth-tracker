package handler

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/internal/pub"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/celoutils/v3"
	"github.com/grassrootseconomics/w3-celo"
)

type faucetGiveHandler struct {
	pub pub.Pub
}

const faucetGiveEventName = "FAUCET_GIVE"

var (
	faucetGiveTopicHash = w3.H("0x26162814817e23ec5035d6a2edc6c422da2da2119e27cfca6be65cc2dc55ca4c")
	faucetGiveEvent     = w3.MustNewEvent("Give(address indexed _recipient, address indexed _token, uint256 _amount)")
	faucetGiveToSig     = w3.MustNewFunc("giveTo(address)", "uint256")
	faucetGimmeSig      = w3.MustNewFunc("gimme()", "uint256")
)

func NewFaucetGiveHandler(pub pub.Pub) *faucetGiveHandler {
	return &faucetGiveHandler{
		pub: pub,
	}
}

func (h *faucetGiveHandler) Name() string {
	return faucetGiveEventName
}

func (h *faucetGiveHandler) HandleLog(ctx context.Context, msg LogMessage) error {
	if msg.Log.Topics[0] == faucetGiveTopicHash {
		var (
			recipient common.Address
			token     common.Address
			amount    big.Int
		)

		if err := faucetGiveEvent.DecodeArgs(msg.Log, &recipient, &token, &amount); err != nil {
			return err
		}

		faucetGiveEvent := event.Event{
			Index:           msg.Log.Index,
			Block:           msg.Log.BlockNumber,
			ContractAddress: msg.Log.Address.Hex(),
			Success:         true,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.Log.TxHash.Hex(),
			TxType:          faucetGiveEventName,
			Payload: map[string]any{
				"recipient": recipient.Hex(),
				"token":     token.Hex(),
				"amount":    amount.String(),
			},
		}

		return h.pub.Send(ctx, faucetGiveEvent)
	}

	return nil
}

func (h *faucetGiveHandler) HandleRevert(ctx context.Context, msg RevertMessage) error {
	if len(msg.InputData) < 8 {
		return nil
	}

	faucetGiveEvent := event.Event{
		Block:           msg.Block,
		ContractAddress: msg.ContractAddress,
		Success:         false,
		Timestamp:       msg.Timestamp,
		TxHash:          msg.TxHash,
		TxType:          faucetGiveEventName,
	}

	switch msg.InputData[:8] {
	case "63e4bff4":
		var to common.Address

		if err := faucetGiveToSig.DecodeArgs(w3.B(msg.InputData), &to); err != nil {
			return err
		}

		faucetGiveEvent.Payload = map[string]any{
			"revertReason": msg.RevertReason,
			"recipient":    to.Hex(),
			"token":        celoutils.ZeroAddress,
			"amount":       "0",
		}

		return h.pub.Send(ctx, faucetGiveEvent)
	case "de82efb4":
		faucetGiveEvent.Payload = map[string]any{
			"revertReason": msg.RevertReason,
			"recipient":    celoutils.ZeroAddress,
			"token":        celoutils.ZeroAddress,
			"amount":       "0",
		}

		return h.pub.Send(ctx, faucetGiveEvent)
	}

	return nil
}
