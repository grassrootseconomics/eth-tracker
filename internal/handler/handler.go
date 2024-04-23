package handler

import (
	"context"

	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/grassrootseconomics/celo-tracker/internal/cache"
	"github.com/grassrootseconomics/celo-tracker/internal/pub"
)

type (
	Handler interface {
		Name() string
		HandleLog(context.Context, LogMessage, pub.Pub) error
		HandleRevert(context.Context, RevertMessage, pub.Pub) error
	}

	LogMessage struct {
		Log       *types.Log
		BlockTime uint64
	}

	RevertMessage struct {
		From            string
		RevertReason    string
		InputData       string
		Block           uint64
		ContractAddress string
		Timestamp       uint64
		TxHash          string
	}
)

func New(cache cache.Cache) []Handler {
	return []Handler{
		&TokenTransferHandler{},
		&PoolSwapHandler{},
		&FaucetGiveHandler{},
		&PoolDepositHandler{},
		&TokenMintHandler{},
		&TokenBurnHandler{},
		&QuoterPriceHandler{},
		&OwnershipHandler{},
		&SealHandler{},
		&IndexAddHandler{
			cache: cache,
		},
		&IndexRemoveHandler{
			cache: cache,
		},
	}
}
