package handlers

import (
	"context"

	"dummy-spot-test-stream-instance/internal/domain"
	"dummy-spot-test-stream-instance/internal/usecase"
	"dummy-spot-test-stream-instance/internal/usecase/ports"
)

type OrderBookHandler struct {
	repository ports.OrderBookRepository
}

func NewOrderBookHandler(repository ports.OrderBookRepository) *OrderBookHandler {
	return &OrderBookHandler{repository: repository}
}

func (handler *OrderBookHandler) Handle(ctx context.Context, packet *usecase.Packet) error {
	if packet == nil || handler.repository == nil {
		return nil
	}

	if command := packet.Payload.Command; command != nil {
		if command.Channel == domain.SubscriptionChannelDepth &&
			command.Action == domain.SubscriptionActionUnsubscribe {
			return handler.repository.Remove(ctx, command.Symbol)
		}
	}

	if depth := packet.Payload.Depth; depth != nil {
		snapshot, err := handler.repository.UpsertDepth(ctx, *depth)
		if err != nil {
			return err
		}
		packet.Payload.Depth = &snapshot
	}

	return nil
}
