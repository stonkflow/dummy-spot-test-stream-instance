package proto

import (
	"encoding/json"
	"fmt"
	"strings"

	"dummy-spot-test-stream-instance/internal/domain"
)

type CommandDecoder struct{}

func NewCommandDecoder() *CommandDecoder {
	return &CommandDecoder{}
}

type commandEnvelope struct {
	RequestID   string       `json:"request_id"`
	PayloadType string       `json:"payload_type"`
	Command     *commandWire `json:"command,omitempty"`
	Action      string       `json:"action,omitempty"`
	Channel     string       `json:"channel,omitempty"`
	Symbol      string       `json:"symbol,omitempty"`
}

type commandWire struct {
	Action  string `json:"action"`
	Channel string `json:"channel"`
	Symbol  string `json:"symbol"`
}

func (decoder *CommandDecoder) Decode(raw []byte) (domain.Payload, error) {
	var envelope commandEnvelope
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return domain.Payload{}, fmt.Errorf("decode command envelope: %w", err)
	}

	if envelope.PayloadType != "" && !strings.EqualFold(envelope.PayloadType, "command") {
		return domain.Payload{}, fmt.Errorf("unexpected payload_type: %s", envelope.PayloadType)
	}

	wire := commandWire{
		Action:  envelope.Action,
		Channel: envelope.Channel,
		Symbol:  envelope.Symbol,
	}
	if envelope.Command != nil {
		wire = *envelope.Command
	}

	action, err := parseAction(wire.Action)
	if err != nil {
		return domain.Payload{}, err
	}
	channel, err := parseChannel(wire.Channel)
	if err != nil {
		return domain.Payload{}, err
	}
	if wire.Symbol == "" {
		return domain.Payload{}, fmt.Errorf("symbol is required")
	}

	command := domain.SubscriptionCommand{
		RequestID: envelope.RequestID,
		Symbol:    wire.Symbol,
		Action:    action,
		Channel:   channel,
	}

	return domain.Payload{
		Command: &command,
	}, nil
}

func parseAction(value string) (domain.SubscriptionAction, error) {
	switch strings.ToUpper(value) {
	case string(domain.SubscriptionActionSubscribe):
		return domain.SubscriptionActionSubscribe, nil
	case string(domain.SubscriptionActionUnsubscribe):
		return domain.SubscriptionActionUnsubscribe, nil
	default:
		return "", fmt.Errorf("unsupported action: %s", value)
	}
}

func parseChannel(value string) (domain.SubscriptionChannel, error) {
	switch strings.ToUpper(value) {
	case string(domain.SubscriptionChannelTrade):
		return domain.SubscriptionChannelTrade, nil
	case string(domain.SubscriptionChannelDepth):
		return domain.SubscriptionChannelDepth, nil
	default:
		return "", fmt.Errorf("unsupported channel: %s", value)
	}
}
