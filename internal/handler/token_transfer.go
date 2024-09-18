package handler

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/pkg/event"
	"github.com/grassrootseconomics/celo-tracker/pkg/router"
	"github.com/grassrootseconomics/celoutils/v3"
	"github.com/grassrootseconomics/w3-celo"
)

const transferEventName = "TOKEN_TRANSFER"

var (
	tokenTransferEvent   = w3.MustNewEvent("Transfer(address indexed _from, address indexed _to, uint256 _value)")
	tokenTransferSig     = w3.MustNewFunc("transfer(address, uint256)", "bool")
	tokenTransferFromSig = w3.MustNewFunc("transferFrom(address, address, uint256)", "bool")

	stables = map[string]bool{
		celoutils.CUSDContractMainnet: true,
		celoutils.CKESContractMainnet: true,
		celoutils.USDTContractMainnet: true,
		celoutils.USDCContractMainnet: true,
	}
)

func HandleTokenTransferLog(hc *HandlerContainer) router.LogHandlerFunc {
	return func(ctx context.Context, lp router.LogPayload, c router.Callback) error {
		var (
			from  common.Address
			to    common.Address
			value big.Int
		)

		if err := tokenTransferEvent.DecodeArgs(lp.Log, &from, &to, &value); err != nil {
			return err
		}

		proceed, err := hc.checkStables(ctx, from.Hex(), to.Hex(), lp.Log.Address.Hex())
		if err != nil {
			return err
		}
		if !proceed {
			return nil
		}

		tokenTransferEvent := event.Event{
			Index:           lp.Log.Index,
			Block:           lp.Log.BlockNumber,
			ContractAddress: lp.Log.Address.Hex(),
			Success:         true,
			Timestamp:       lp.Timestamp,
			TxHash:          lp.Log.TxHash.Hex(),
			TxType:          transferEventName,
			Payload: map[string]any{
				"from":  from.Hex(),
				"to":    to.Hex(),
				"value": value.String(),
			},
		}

		return c(ctx, tokenTransferEvent)
	}
}

func HandleTokenTransferInputData(hc *HandlerContainer) router.InputDataHandlerFunc {
	return func(ctx context.Context, idp router.InputDataPayload, c router.Callback) error {
		tokenTransferEvent := event.Event{
			Block:           idp.Block,
			ContractAddress: idp.ContractAddress,
			Success:         false,
			Timestamp:       idp.Timestamp,
			TxHash:          idp.TxHash,
			TxType:          transferEventName,
		}

		switch idp.InputData[:8] {
		case "a9059cbb":
			var (
				to    common.Address
				value big.Int
			)

			if err := tokenTransferSig.DecodeArgs(w3.B(idp.InputData), &to, &value); err != nil {
				return err
			}

			proceed, err := hc.checkStables(ctx, idp.From, to.Hex(), idp.ContractAddress)
			if err != nil {
				return err
			}
			if !proceed {
				return nil
			}

			tokenTransferEvent.Payload = map[string]any{
				"from":  idp.From,
				"to":    to.Hex(),
				"value": value.String(),
			}

			return c(ctx, tokenTransferEvent)
		case "23b872dd":
			var (
				from  common.Address
				to    common.Address
				value big.Int
			)

			if err := tokenTransferFromSig.DecodeArgs(w3.B(idp.InputData), &from, &to, &value); err != nil {
				return err
			}

			proceed, err := hc.checkStables(ctx, from.Hex(), to.Hex(), idp.ContractAddress)
			if err != nil {
				return err
			}
			if !proceed {
				return nil
			}

			tokenTransferEvent.Payload = map[string]any{
				"from":  from.Hex(),
				"to":    to.Hex(),
				"value": value.String(),
			}

			return c(ctx, tokenTransferEvent)
		}

		return nil
	}
}

func (hc *HandlerContainer) checkStables(ctx context.Context, from string, to string, contractAddress string) (bool, error) {
	_, ok := stables[contractAddress]
	if !ok {
		return true, nil
	}

	// TODO: Pipeline this check on Redis with a new method
	fromExists, err := hc.cache.Exists(ctx, from)
	if err != nil {
		return false, err
	}

	toExists, err := hc.cache.Exists(ctx, to)
	if err != nil {
		return false, err
	}

	return fromExists || toExists, nil
}