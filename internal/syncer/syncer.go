package syncer

import (
	"log/slog"

	"github.com/bits-and-blooms/bitset"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/celo-org/celo-blockchain/ethclient"
	"github.com/ef-ds/deque/v2"
	"github.com/grassrootseconomics/celo-events/internal/chain"
	"github.com/grassrootseconomics/celo-events/internal/stats"
)

const (
	blockBatchSize = 100
)

type (
	SyncerOpts struct {
		WebSocketEndpoint string
		BatchQueue        *deque.Deque[uint64]
		BlocksQueue       *deque.Deque[types.Block]
		Chain             *chain.Chain
		Logg              *slog.Logger
		Stats             *stats.Stats
		// replace with db
		Db *bitset.BitSet
	}

	Syncer struct {
		batchQueue  *deque.Deque[uint64]
		blocksQueue *deque.Deque[types.Block]
		chain       *chain.Chain
		logg        *slog.Logger
		stats       *stats.Stats
		ethClient   *ethclient.Client
		db          *bitset.BitSet
	}
)

func NewSyncer(o SyncerOpts) (*Syncer, error) {
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
		db:          o.Db,
	}, nil
}
