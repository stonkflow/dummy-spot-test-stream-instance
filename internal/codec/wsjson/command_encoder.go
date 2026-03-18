package wsjson

import (
	"encoding/json"
	"fmt"
	"strings"

	"dummy-spot-test-stream-instance/internal/domain"
)

type CommandEncoder struct{}

func NewCommandEncoder() *CommandEncoder {
	return &CommandEncoder{}
}

type commandPayload struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	ID     string   `json:"id,omitempty"`
}

func (encoder *CommandEncoder) Encode(payload domain.Payload) ([]byte, error) {
	command, err := commandFromPayload(payload)
	if err != nil {
		return nil, err
	}

	method := string(command.Action)
	if method == "" {
		return nil, fmt.Errorf("command action is required")
	}
	if command.Symbol == "" {
		return nil, fmt.Errorf("command symbol is required")
	}

	streamName, err := streamName(command.Symbol, command.Channel)
	if err != nil {
		return nil, err
	}

	encoded, err := json.Marshal(commandPayload{
		Method: method,
		Params: []string{streamName},
		ID:     command.RequestID,
	})
	if err != nil {
		return nil, fmt.Errorf("encode ws command: %w", err)
	}

	return encoded, nil
}

func commandFromPayload(payload domain.Payload) (domain.SubscriptionCommand, error) {
	if payload.Command == nil {
		return domain.SubscriptionCommand{}, fmt.Errorf("subscription command payload is required")
	}
	return *payload.Command, nil
}

func streamName(symbol string, channel domain.SubscriptionChannel) (string, error) {
	switch channel {
	case domain.SubscriptionChannelTrade:
		return strings.ToLower(symbol) + "@trade", nil
	case domain.SubscriptionChannelDepth:
		return strings.ToLower(symbol) + "@depth", nil
	default:
		return "", fmt.Errorf("unsupported subscription channel: %s", channel)
	}
}
