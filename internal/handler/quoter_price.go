package handler

import (
	"context"
	"math/big"

	"github.com/celo-org/celo-blockchain/common"
	"github.com/grassrootseconomics/celo-tracker/internal/event"
	"github.com/grassrootseconomics/celo-tracker/internal/pub"
	"github.com/grassrootseconomics/w3-celo"
)

type (
	QuoterPriceHandler struct {
		topicHash common.Hash
		event     *w3.Event
	}
)

const (
	quoterPriceEventName = "QUOTER_PRICE_INDEX_UPDATED"
)

var (
	quoterPriceTopicHash = w3.H("0xdb9ce1a76955721ca61ac50cd1b87f9ab8620325c8619a62192c2dc7871d56b1")
	quoterPriceEvent     = w3.MustNewEvent("PriceIndexUpdated(address _tokenAddress, uint256 _exchangeRate)")
	quoterPriceToSig     = w3.MustNewFunc("setPriceIndexValue(address, uint256)", "uint256")
)

func (h *QuoterPriceHandler) Name() string {
	return quoterPriceEventName
}

func (h *QuoterPriceHandler) HandleLog(ctx context.Context, msg LogMessage, pub pub.Pub) error {
	if msg.Log.Topics[0] == quoterPriceTopicHash {
		var (
			token        common.Address
			exchangeRate big.Int
		)

		if err := quoterPriceEvent.DecodeArgs(msg.Log, &token, &exchangeRate); err != nil {
			return err
		}

		quoterPriceEvent := event.Event{
			Block:           msg.Log.BlockNumber,
			ContractAddress: msg.Log.Address.Hex(),
			Success:         true,
			Timestamp:       msg.BlockTime,
			TxHash:          msg.Log.TxHash.Hex(),
			TxType:          quoterPriceEventName,
			Payload: map[string]any{
				"token":        token.Hex(),
				"exchangeRate": exchangeRate.String(),
			},
		}

		return pub.Send(ctx, quoterPriceEvent)
	}

	return nil
}

func (h *QuoterPriceHandler) HandleRevert(ctx context.Context, msg RevertMessage, pub pub.Pub) error {
	if len(msg.InputData) < 8 {
		return nil
	}

	switch msg.InputData[:8] {
	case "ebc59dff":
		var (
			token        common.Address
			exchangeRate big.Int
		)

		if err := quoterPriceToSig.DecodeArgs(w3.B(msg.InputData), &token, &exchangeRate); err != nil {
			return err
		}

		quoterPriceEvent := event.Event{
			Block:           msg.Block,
			ContractAddress: msg.ContractAddress,
			Success:         false,
			Timestamp:       msg.Timestamp,
			TxHash:          msg.TxHash,
			TxType:          quoterPriceEventName,
			Payload: map[string]any{
				"revertReason": msg.RevertReason,
				"token":        token.Hex(),
				"exchangeRate": exchangeRate.String(),
			},
		}

		return pub.Send(ctx, quoterPriceEvent)
	}

	return nil
}
