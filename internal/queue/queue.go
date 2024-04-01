package queue

import (
	"github.com/celo-org/celo-blockchain/core/types"
	"github.com/ef-ds/deque/v2"
)

type (
	Queue struct {
		BlocksQueue *deque.Deque[types.Block]
		BatchQueue  *deque.Deque[uint64]
	}
)
