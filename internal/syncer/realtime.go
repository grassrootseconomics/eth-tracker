package syncer

import (
	"context"
	"errors"
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
	resubscribeInterval = 2 * time.Second
)

func (s *Syncer) Start() {
	s.realtimeSub = event.ResubscribeErr(resubscribeInterval, s.resubscribeFn())
}

func (s *Syncer) Stop() {
	if s.realtimeSub != nil {
		s.realtimeSub.Unsubscribe()
	}
}

func (s *Syncer) receiveRealtimeBlocks(ctx context.Context, fn BlockQueueFn) (celo.Subscription, error) {
	newHeadersReceiver := make(chan *types.Header, 1)
	sub, err := s.ethClient.SubscribeNewHead(ctx, newHeadersReceiver)
	if err != nil {
		return nil, err
	}
	s.logg.Info("realtime syncer connected to ws endpoint")

	return event.NewSubscription(func(quit <-chan struct{}) error {
		eventsCtx, eventsCancel := context.WithCancel(context.Background())
		defer sub.Unsubscribe()
		defer eventsCancel()

		go func() {
			select {
			case <-quit:
				s.logg.Info("realtime syncer stopping")
				eventsCancel()
			case <-eventsCtx.Done():
				return
			}
		}()

		for {
			select {
			case header := <-newHeadersReceiver:
				if err := fn(eventsCtx, header.Number.Uint64()); err != nil {
					if !errors.Is(err, context.Canceled) {
						s.logg.Error("realtime block queuer error", "error", err)
					}
				}
			case <-eventsCtx.Done():
				s.logg.Info("realtime syncer shutting down")
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
		if !errors.Is(err, context.Canceled) {
			return fmt.Errorf("block %d error: %v", blockNumber, err)
		}
	}

	s.blockWorker.Submit(func() {
		if err := s.blockProcessor.ProcessBlock(context.Background(), block); err != nil {
			s.logg.Error("block processor error", "source", "realtime", "block", blockNumber, "error", err)
		}
	})

	if err := s.db.SetUpperBound(blockNumber); err != nil {
		return err
	}

	return nil
}

func (s *Syncer) resubscribeFn() event.ResubscribeErrFunc {
	return func(ctx context.Context, err error) (event.Subscription, error) {
		if err != nil {
			s.logg.Error("resubscribing after failed subscription", "error", err)
		}
		return s.receiveRealtimeBlocks(ctx, s.queueRealtimeBlock)
	}
}
