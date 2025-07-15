package handler

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/grassrootseconomics/eth-tracker/pkg/event"
	"github.com/grassrootseconomics/eth-tracker/pkg/router"
	"github.com/lmittmann/w3"
)

const limitSetEventName = "LIMIT_SET"

var (
	limitSetEvent = w3.MustNewEvent("LimitSet(address indexed token, address indexed holder, uint256 value)")
	limitSetSig   = w3.MustNewFunc("setLimitFor(address, address, uint256)", "")
)

func HandleLimitSetLog(hc *HandlerContainer) router.LogHandlerFunc {
	return func(ctx context.Context, lp router.LogPayload, c router.Callback) error {
		var (
			token  common.Address
			holder common.Address
			value  big.Int
		)

		if err := limitSetEvent.DecodeArgs(lp.Log, &token, &holder, &value); err != nil {
			return err
		}

		proceed, err := hc.checkWithinNetwork(ctx, lp.Log.Address.Hex(), token.Hex(), holder.Hex())
		if err != nil {
			return err
		}
		if !proceed {
			return nil
		}

		limitSetEvent := event.Event{
			Index:           lp.Log.Index,
			Block:           lp.Log.BlockNumber,
			ContractAddress: lp.Log.Address.Hex(),
			Success:         true,
			Timestamp:       lp.Timestamp,
			TxHash:          lp.Log.TxHash.Hex(),
			TxType:          limitSetEventName,
			Payload: map[string]any{
				"token":  token.Hex(),
				"holder": holder.Hex(),
				"value":  value.String(),
			},
		}

		return c(ctx, limitSetEvent)
	}
}

func HandleLimitSetInputData(hc *HandlerContainer) router.InputDataHandlerFunc {
	return func(ctx context.Context, idp router.InputDataPayload, c router.Callback) error {
		var (
			token  common.Address
			holder common.Address
			value  big.Int
		)

		if err := limitSetSig.DecodeArgs(w3.B(idp.InputData), &token, &holder, &value); err != nil {
			return err
		}

		proceed, err := hc.checkWithinNetwork(ctx, idp.ContractAddress, token.Hex(), holder.Hex())
		if err != nil {
			return err
		}
		if !proceed {
			return nil
		}

		limitSetEvent := event.Event{
			Block:           idp.Block,
			ContractAddress: idp.ContractAddress,
			Success:         false,
			Timestamp:       idp.Timestamp,
			TxHash:          idp.TxHash,
			TxType:          limitSetEventName,
			Payload: map[string]any{
				"token":  token.Hex(),
				"holder": holder.Hex(),
				"value":  value.String(),
			},
		}

		return c(ctx, limitSetEvent)
	}
}
