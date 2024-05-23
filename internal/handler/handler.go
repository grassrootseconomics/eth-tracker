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
		HandleLog(context.Context, LogMessage) error
		HandleRevert(context.Context, RevertMessage) error
	}

	HandlerPipeline []Handler

	LogMessage struct {
		Log       *types.Log
		Timestamp uint64
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

func New(pub pub.Pub, cache cache.Cache) HandlerPipeline {
	return []Handler{
		NewTokenTransferHandler(pub),
		NewPoolSwapHandler(pub),
		NewFaucetGiveHandler(pub),
		NewPoolDepositHandler(pub),
		NewTokenMintHandler(pub),
		NewTokenBurnHandler(pub),
		NewQuoterPriceHandler(pub),
		NewOwnershipHandler(pub),
		NewSealHandler(pub),
		NewIndexAddHandler(pub, cache),
		NewIndexRemoveHandler(pub, cache),
	}
}
