package syncer

import (
	"context"
	"log/slog"

	"github.com/celo-org/celo-blockchain"
	"github.com/celo-org/celo-blockchain/ethclient"
	"github.com/gammazero/workerpool"
	"github.com/grassrootseconomics/celo-tracker/internal/db"
	"github.com/grassrootseconomics/celo-tracker/internal/processor"
	"github.com/grassrootseconomics/celo-tracker/internal/stats"
	"github.com/grassrootseconomics/celo-tracker/pkg/chain"
)

type (
	SyncerOpts struct {
		BlockWorker       *workerpool.WorkerPool
		BlockProcessor    *processor.Processor
		Chain             *chain.Chain
		DB                *db.DB
		Logg              *slog.Logger
		StartBlock        int64
		Stats             *stats.Stats
		WebSocketEndpoint string
	}

	Syncer struct {
		blockWorker    *workerpool.WorkerPool
		blockProcessor *processor.Processor
		chain          *chain.Chain
		db             *db.DB
		ethClient      *ethclient.Client
		logg           *slog.Logger
		quit           chan struct{}
		realtimeSub    celo.Subscription
		stats          *stats.Stats
	}
)

func New(o SyncerOpts) (*Syncer, error) {
	latestBlock, err := o.Chain.GetLatestBlock(context.Background())
	if err != nil {
		return nil, err
	}

	lowerBound, err := o.DB.GetLowerBound()
	if err != nil {
		return nil, err
	}
	if lowerBound == 0 {
		if o.StartBlock > 0 {
			if err := o.DB.SetLowerBound(uint64(o.StartBlock)); err != nil {
				return nil, err
			}
		} else {
			if err := o.DB.SetLowerBound(latestBlock); err != nil {
				return nil, err
			}
		}
	}

	if err := o.DB.SetUpperBound(latestBlock); err != nil {
		return nil, err
	}

	ethClient, err := ethclient.Dial(o.WebSocketEndpoint)
	if err != nil {
		return nil, err
	}

	return &Syncer{
		blockWorker:    o.BlockWorker,
		blockProcessor: o.BlockProcessor,
		chain:          o.Chain,
		db:             o.DB,
		ethClient:      ethClient,
		logg:           o.Logg,
		quit:           make(chan struct{}),
		stats:          o.Stats,
	}, nil
}
