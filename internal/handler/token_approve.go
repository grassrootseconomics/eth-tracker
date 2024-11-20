package handler

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/eth-tracker/pkg/event"
	"github.com/grassrootseconomics/eth-tracker/pkg/router"
	"github.com/grassrootseconomics/w3-celo"
)

const approveEventName = "TOKEN_APPROVE"

var (
	tokenApproveEvent = w3.MustNewEvent("Approval(address indexed _owner, address indexed _spender, uint256 _value)")
	tokenApproveToSig = w3.MustNewFunc("approve(address, uint256)", "bool")
)

func HandleTokenApproveLog(hc *HandlerContainer) router.LogHandlerFunc {
	return func(ctx context.Context, lp router.LogPayload, c router.Callback) error {
		var (
			owner   common.Address
			spender common.Address
			value   big.Int
		)

		if err := tokenApproveEvent.DecodeArgs(lp.Log, &owner, &spender, &value); err != nil {
			return err
		}

		proceed, err := hc.checkWithinNetwork(ctx, lp.Log.Address.Hex(), owner.Hex(), spender.Hex())
		if err != nil {
			return err
		}
		if !proceed {
			return nil
		}

		tokenApproveEvent := event.Event{
			Index:           lp.Log.Index,
			Block:           lp.Log.BlockNumber,
			ContractAddress: lp.Log.Address.Hex(),
			Success:         true,
			Timestamp:       lp.Timestamp,
			TxHash:          lp.Log.TxHash.Hex(),
			TxType:          approveEventName,
			Payload: map[string]any{
				"owner":   owner.Hex(),
				"spender": spender.Hex(),
				"value":   value.String(),
			},
		}

		return c(ctx, tokenApproveEvent)
	}
}

func HandleTokenApproveInputData(hc *HandlerContainer) router.InputDataHandlerFunc {
	return func(ctx context.Context, idp router.InputDataPayload, c router.Callback) error {
		var (
			spender common.Address
			value   big.Int
		)

		if err := tokenApproveToSig.DecodeArgs(w3.B(idp.InputData), &spender, &value); err != nil {
			return err
		}

		proceed, err := hc.checkWithinNetwork(ctx, idp.ContractAddress, idp.From, spender.Hex())
		if err != nil {
			return err
		}
		if !proceed {
			return nil
		}

		tokenApproveEvent := event.Event{
			Block:           idp.Block,
			ContractAddress: idp.ContractAddress,
			Success:         false,
			Timestamp:       idp.Timestamp,
			TxHash:          idp.TxHash,
			TxType:          approveEventName,
			Payload: map[string]any{
				"owner":   idp.From,
				"spender": spender.Hex(),
				"value":   value.String(),
			},
		}

		return c(ctx, tokenApproveEvent)
	}
}
