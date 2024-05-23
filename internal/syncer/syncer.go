package syncer

import (
	"context"
	"log/slog"

	"github.com/celo-org/celo-blockchain"
	"github.com/celo-org/celo-blockchain/ethclient"
	"github.com/grassrootseconomics/celo-tracker/internal/chain"
	"github.com/grassrootseconomics/celo-tracker/internal/db"
	"github.com/grassrootseconomics/celo-tracker/internal/queue"
	"github.com/grassrootseconomics/celo-tracker/internal/stats"
)

type (
	SyncerOpts struct {
		DB                db.DB
		Chain             chain.Chain
		Logg              *slog.Logger
		Queue             *queue.Queue
		Stats             *stats.Stats
		StartBlock        int64
		WebSocketEndpoint string
	}

	Syncer struct {
		db          db.DB
		ethClient   *ethclient.Client
		logg        *slog.Logger
		realtimeSub celo.Subscription
		stats       *stats.Stats
		queue       *queue.Queue
		stopCh      chan struct{}
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
	o.Stats.SetLatestBlock(latestBlock)

	ethClient, err := ethclient.Dial(o.WebSocketEndpoint)
	if err != nil {
		return nil, err
	}

	return &Syncer{
		db:        o.DB,
		ethClient: ethClient,
		logg:      o.Logg,
		stats:     o.Stats,
		queue:     o.Queue,
		stopCh:    make(chan struct{}),
	}, nil
}
