package wsjson

import (
	"encoding/json"
	"fmt"
	"strings"

	"dummy-spot-test-stream-instance/internal/domain"
)

type EventDecoder struct{}

func NewEventDecoder() *EventDecoder {
	return &EventDecoder{}
}

type wsEventEnvelope struct {
	EventType string     `json:"e"`
	Symbol    string     `json:"s"`
	EventTime int64      `json:"E"`
	FinalID   int64      `json:"u"`
	Bids      [][]string `json:"b"`
	Asks      [][]string `json:"a"`
	Price     string     `json:"p"`
	Quantity  string     `json:"q"`
	TradeID   int64      `json:"t"`
	Maker     bool       `json:"m"`

	Type string          `json:"type"`
	TS   int64           `json:"ts"`
	Data json.RawMessage `json:"data"`
}

type normalizedTrade struct {
	TradeID      int64  `json:"tradeId"`
	Price        string `json:"price"`
	Quantity     string `json:"quantity"`
	IsBuyerMaker bool   `json:"isBuyerMaker"`
}

type normalizedDepth struct {
	LastUpdateID int64        `json:"lastUpdateId"`
	Bids         [][]string   `json:"bids"`
	Asks         [][]string   `json:"asks"`
	Book         *depthInline `json:"book,omitempty"`
}

type depthInline struct {
	Bids [][]string `json:"bids"`
	Asks [][]string `json:"asks"`
}

func (decoder *EventDecoder) Decode(raw []byte) (domain.Payload, error) {
	var envelope wsEventEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return domain.Payload{}, fmt.Errorf("decode ws event: %w", err)
	}

	if envelope.Type != "" {
		return decodeNormalized(envelope)
	}

	switch strings.ToLower(envelope.EventType) {
	case "trade":
		trade := domain.TradeEvent{
			Symbol:       envelope.Symbol,
			EventTimeMS:  envelope.EventTime,
			TradeID:      envelope.TradeID,
			Price:        envelope.Price,
			Quantity:     envelope.Quantity,
			IsBuyerMaker: envelope.Maker,
		}
		return domain.Payload{Trade: &trade}, nil
	case "depthupdate", "depth":
		depth := domain.DepthEvent{
			Symbol:       envelope.Symbol,
			EventTimeMS:  envelope.EventTime,
			LastUpdateID: envelope.FinalID,
			Bids:         toDepthLevels(envelope.Bids),
			Asks:         toDepthLevels(envelope.Asks),
		}
		return domain.Payload{Depth: &depth}, nil
	default:
		return domain.Payload{}, fmt.Errorf("unsupported ws event type: %s", envelope.EventType)
	}
}

func decodeNormalized(envelope wsEventEnvelope) (domain.Payload, error) {
	switch strings.ToLower(envelope.Type) {
	case "trade":
		var trade normalizedTrade
		if err := json.Unmarshal(envelope.Data, &trade); err != nil {
			return domain.Payload{}, fmt.Errorf("decode normalized trade: %w", err)
		}
		event := domain.TradeEvent{
			Symbol:       envelope.Symbol,
			EventTimeMS:  envelope.TS,
			TradeID:      trade.TradeID,
			Price:        trade.Price,
			Quantity:     trade.Quantity,
			IsBuyerMaker: trade.IsBuyerMaker,
		}
		return domain.Payload{Trade: &event}, nil
	case "depth":
		var depth normalizedDepth
		if err := json.Unmarshal(envelope.Data, &depth); err != nil {
			return domain.Payload{}, fmt.Errorf("decode normalized depth: %w", err)
		}
		bids := depth.Bids
		asks := depth.Asks
		if depth.Book != nil {
			if len(depth.Book.Bids) > 0 {
				bids = depth.Book.Bids
			}
			if len(depth.Book.Asks) > 0 {
				asks = depth.Book.Asks
			}
		}
		event := domain.DepthEvent{
			Symbol:       envelope.Symbol,
			EventTimeMS:  envelope.TS,
			LastUpdateID: depth.LastUpdateID,
			Bids:         toDepthLevels(bids),
			Asks:         toDepthLevels(asks),
		}
		return domain.Payload{Depth: &event}, nil
	default:
		return domain.Payload{}, fmt.Errorf("unsupported normalized ws event type: %s", envelope.Type)
	}
}

func toDepthLevels(levels [][]string) []domain.DepthLevel {
	result := make([]domain.DepthLevel, 0, len(levels))
	for _, level := range levels {
		if len(level) < 2 {
			continue
		}
		result = append(result, domain.DepthLevel{
			Price:    level[0],
			Quantity: level[1],
		})
	}
	return result
}
