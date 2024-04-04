package syncer

import (
	"context"
	"fmt"
	"time"

	"github.com/celo-org/celo-blockchain"
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/celo-org/celo-blockchain/event"
)

type (
	BlockQueueFn func(context.Context, uint64) error
)

const (
	resubscribeInterval = 5 * time.Second
)

// func (s *Syncer) StartRealtimeSyncer(ctx context.Context) error {
// 	newHeadersReceiver := make(chan *types.Header, 1)

// 	sub := event.ResubscribeErr(resubscribeInterval, func(ctx context.Context, err error) (event.Subscription, error) {
// 		if err != nil {
// 			s.logg.Error("realtime syncer resubscribe error", "error", err)
// 		}
// 		return s.ethClient.SubscribeNewHead(ctx, newHeadersReceiver)
// 	})
// 	defer sub.Unsubscribe()

// 	for {
// 		select {
// 		case <-ctx.Done():
// 			s.logg.Info("realtime syncer shutting down")
// 			return nil
// 		case header := <-newHeadersReceiver:
// 			blockNumber := header.Number.Uint64()
// 			block, err := s.chain.GetBlock(context.Background(), blockNumber)
// 			if err != nil {
// 				s.logg.Error("realtime block fetcher error", "block", blockNumber, "error", err)
// 			}
// 			s.blocksQueue.PushFront(block)
// 		}
// 	}
// }

func (s *Syncer) StartRealtime() {
	s.realtimeSub = event.ResubscribeErr(resubscribeInterval, s.resubscribeFn())
}

func (s *Syncer) StopRealtime() {
	s.realtimeSub.Unsubscribe()
	s.realtimeSub = nil
}

func (s *Syncer) receiveRealtimeBlocks(ctx context.Context, fn BlockQueueFn) (celo.Subscription, error) {
	newHeadersReceiver := make(chan *types.Header, 10)
	sub, err := s.ethClient.SubscribeNewHead(ctx, newHeadersReceiver)
	if err != nil {
		return nil, err
	}

	return event.NewSubscription(func(quit <-chan struct{}) error {
		eventsCtx, eventsCancel := context.WithCancel(context.Background())
		defer sub.Unsubscribe()
		defer eventsCancel()

		go func() {
			select {
			case <-quit:
				eventsCancel()
			case <-eventsCtx.Done():
				return
			}
		}()

		for {
			select {
			case header := <-newHeadersReceiver:
				s.logg.Debug("received block", "block", header.Number.Uint64())
				if err := fn(eventsCtx, header.Number.Uint64()); err != nil {
					s.logg.Error("realtime block queuer error", "error", err)
				}
			case <-eventsCtx.Done():
				return nil
			case err := <-sub.Err():
				return err
			}
		}
	}), nil
}

func (s *Syncer) queueRealtimeBlock(ctx context.Context, blockNumber uint64) error {
	block, err := s.chain.GetBlock(ctx, blockNumber)
	if err != nil {
		return fmt.Errorf("block %d error: %v", blockNumber, err)
	}
	s.blocksQueue.PushFront(block)
	s.logg.Debug("queued block", "block", blockNumber)
	return nil
}

func (s *Syncer) resubscribeFn() event.ResubscribeErrFunc {
	return func(ctx context.Context, err error) (event.Subscription, error) {
		if err != nil {
			s.logg.Error("resubscribing after failed suibscription", "error", err)
		}
		return s.receiveRealtimeBlocks(ctx, s.queueRealtimeBlock)
	}
}
