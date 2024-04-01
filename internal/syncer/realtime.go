package syncer

import (
	"context"
	"time"

	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/celo-org/celo-blockchain/event"
)

const (
	resubscribeInterval = 5 * time.Second
)

func (s *Syncer) StartRealtimeSyncer(ctx context.Context) error {
	newHeadersReceiver := make(chan *types.Header, 1)

	sub := event.ResubscribeErr(resubscribeInterval, func(ctx context.Context, err error) (event.Subscription, error) {
		if err != nil {
			s.logg.Error("realtime syncer resubscribe error", "error", err)
		}
		return s.ethClient.SubscribeNewHead(ctx, newHeadersReceiver)
	})
	defer sub.Unsubscribe()

	for {
		select {
		case <-ctx.Done():
			s.logg.Info("realtime syncer shutting down")
			return nil
		case header := <-newHeadersReceiver:
			blockNumber := header.Number.Uint64()
			block, err := s.chain.GetBlock(context.Background(), blockNumber)
			if err != nil {
				s.logg.Error("realtime block fetcher error", "block", blockNumber, "error", err)
			}
			s.blocksQueue.PushFront(block)
		}
	}
}
