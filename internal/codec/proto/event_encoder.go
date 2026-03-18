package proto

import (
	"encoding/json"
	"fmt"

	"dummy-spot-test-stream-instance/internal/domain"
)

type EventEncoder struct{}

func NewEventEncoder() *EventEncoder {
	return &EventEncoder{}
}

type eventEnvelope struct {
	RequestID   string          `json:"request_id,omitempty"`
	TSMS        int64           `json:"ts_ms,omitempty"`
	PayloadType string          `json:"payload_type"`
	Trade       *tradeWire      `json:"trade,omitempty"`
	Depth       *depthWire      `json:"depth,omitempty"`
	Symbol      string          `json:"symbol,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
}

type tradeWire struct {
	TradeID      int64  `json:"trade_id"`
	Price        string `json:"price"`
	Quantity     string `json:"quantity"`
	IsBuyerMaker bool   `json:"is_buyer_maker"`
}

type depthWire struct {
	LastUpdateID int64       `json:"last_update_id"`
	Bids         []levelWire `json:"bids"`
	Asks         []levelWire `json:"asks"`
}

type levelWire struct {
	Price    string `json:"price"`
	Quantity string `json:"quantity"`
}

func (encoder *EventEncoder) Encode(payload domain.Payload) ([]byte, error) {
	envelope, err := eventEnvelopeFromPayload(payload)
	if err != nil {
		return nil, err
	}

	encoded, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("encode event envelope: %w", err)
	}
	return encoded, nil
}

func eventEnvelopeFromPayload(payload domain.Payload) (eventEnvelope, error) {
	if payload.Trade != nil {
		trade := payload.Trade
		return eventEnvelope{
			RequestID:   trade.RequestID,
			TSMS:        trade.EventTimeMS,
			PayloadType: "trade",
			Symbol:      trade.Symbol,
			Trade: &tradeWire{
				TradeID:      trade.TradeID,
				Price:        trade.Price,
				Quantity:     trade.Quantity,
				IsBuyerMaker: trade.IsBuyerMaker,
			},
		}, nil
	}

	if payload.Depth != nil {
		depth := payload.Depth
		return eventEnvelope{
			RequestID:   depth.RequestID,
			TSMS:        depth.EventTimeMS,
			PayloadType: "depth",
			Symbol:      depth.Symbol,
			Depth: &depthWire{
				LastUpdateID: depth.LastUpdateID,
				Bids:         toLevelWire(depth.Bids),
				Asks:         toLevelWire(depth.Asks),
			},
		}, nil
	}

	if payload.Command != nil {
		return eventEnvelope{}, fmt.Errorf("command payload cannot be encoded as event")
	}

	return eventEnvelope{}, fmt.Errorf("payload is empty")
}

func toLevelWire(levels []domain.DepthLevel) []levelWire {
	result := make([]levelWire, 0, len(levels))
	for _, level := range levels {
		result = append(result, levelWire{
			Price:    level.Price,
			Quantity: level.Quantity,
		})
	}
	return result
}
