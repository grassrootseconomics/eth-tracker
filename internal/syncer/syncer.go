package syncer

import (
	"errors"
	"log/slog"

	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/celo-org/celo-blockchain/ethclient"
	"github.com/ef-ds/deque/v2"
	"github.com/grassrootseconomics/celo-tracker/internal/chain"
	"github.com/grassrootseconomics/celo-tracker/internal/db"
	"github.com/grassrootseconomics/celo-tracker/internal/stats"
)

type (
	SyncerOpts struct {
		WebSocketEndpoint string
		BatchQueue        *deque.Deque[uint64]
		BlocksQueue       *deque.Deque[types.Block]
		Chain             *chain.Chain
		Logg              *slog.Logger
		Stats             *stats.Stats
		DB                *db.DB
		InitialLowerBound uint64
	}

	Syncer struct {
		batchQueue        *deque.Deque[uint64]
		blocksQueue       *deque.Deque[types.Block]
		chain             *chain.Chain
		logg              *slog.Logger
		stats             *stats.Stats
		ethClient         *ethclient.Client
		db                *db.DB
		initialLowerBound uint64
	}
)

func New(o SyncerOpts) (*Syncer, error) {
	if o.InitialLowerBound == 0 {
		return nil, errors.New("initial lower bound not set")
	}

	ethClient, err := ethclient.Dial(o.WebSocketEndpoint)
	if err != nil {
		return nil, err
	}

	return &Syncer{
		batchQueue:        o.BatchQueue,
		blocksQueue:       o.BlocksQueue,
		chain:             o.Chain,
		logg:              o.Logg,
		stats:             o.Stats,
		ethClient:         ethClient,
		db:                o.DB,
		initialLowerBound: o.InitialLowerBound,
	}, nil
}
