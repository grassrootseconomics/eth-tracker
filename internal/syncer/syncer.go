package syncer

import (
	"context"
	"log/slog"

	"github.com/celo-org/celo-blockchain"
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
		EnableHistorical  bool
		StartBlock        uint64
		BatchQueue        *deque.Deque[uint64]
		BlocksQueue       *deque.Deque[types.Block]
		Chain             *chain.Chain
		Logg              *slog.Logger
		Stats             *stats.Stats
		DB                *db.DB
	}

	Syncer struct {
		batchQueue  *deque.Deque[uint64]
		blocksQueue *deque.Deque[types.Block]
		chain       *chain.Chain
		logg        *slog.Logger
		stats       *stats.Stats
		ethClient   *ethclient.Client
		db          *db.DB
		quit        chan struct{}
		startBlock  uint64
		realtimeSub celo.Subscription
	}
)

func New(o SyncerOpts) (*Syncer, error) {
	if o.EnableHistorical {
		latestBlock, err := o.Chain.GetLatestBlock(context.Background())
		if err != nil {
			return nil, err
		}

		if o.StartBlock == 0 {
			o.StartBlock = latestBlock
		}

		if err := o.DB.SetLowerBound(o.StartBlock); err != nil {
			return nil, err
		}

		if err := o.DB.SetUpperBound(latestBlock); err != nil {
			return nil, err
		}
	}

	ethClient, err := ethclient.Dial(o.WebSocketEndpoint)
	if err != nil {
		return nil, err
	}

	return &Syncer{
		batchQueue:  o.BatchQueue,
		blocksQueue: o.BlocksQueue,
		chain:       o.Chain,
		logg:        o.Logg,
		stats:       o.Stats,
		ethClient:   ethClient,
		db:          o.DB,
		quit:        make(chan struct{}),
		startBlock:  o.StartBlock,
	}, nil
}
