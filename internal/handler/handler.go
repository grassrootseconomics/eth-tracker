package handler

import (
	"context"

	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/grassrootseconomics/celo-tracker/internal/emitter"
)

type (
	Handler interface {
		HandleLog(context.Context, LogMessage, emitter.Emitter) error
		HandleRevert(context.Context, RevertMessage, emitter.Emitter) error
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

	Event struct {
		Block           uint64         `json:"block"`
		ContractAddress string         `json:"contractAddress"`
		Success         bool           `json:"success"`
		Timestamp       uint64         `json:"timestamp"`
		TxHash          string         `json:"transactionHash"`
		TxType          string         `json:"transactionType"`
		Payload         map[string]any `json:"payload"`
	}
)

func New() []Handler {
	return []Handler{
		&TokenTransferHandler{},
	}
}
