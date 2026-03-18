package noop

import (
	"context"

	"dummy-spot-test-stream-instance/internal/domain"
)

type CommandConsumer struct{}

func NewCommandConsumer() *CommandConsumer {
	return &CommandConsumer{}
}

func (c *CommandConsumer) Receive(ctx context.Context) ([]byte, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func (c *CommandConsumer) Close() error {
	return nil
}

type EventProducer struct{}

func NewEventProducer() *EventProducer {
	return &EventProducer{}
}

func (p *EventProducer) Send(_ context.Context, _ []byte, _ []byte) error {
	return nil
}

func (p *EventProducer) Close() error {
	return nil
}

type WSClient struct{}

func NewWSClient() *WSClient {
	return &WSClient{}
}

func (c *WSClient) Send(_ context.Context, _ []byte) error {
	return nil
}

func (c *WSClient) Receive(ctx context.Context) ([]byte, error) {
	<-ctx.Done()
	return nil, ctx.Err()
}

func (c *WSClient) Close() error {
	return nil
}

type OrderBookRepository struct{}

func NewOrderBookRepository() *OrderBookRepository {
	return &OrderBookRepository{}
}

func (s *OrderBookRepository) UpsertDepth(_ context.Context, depth domain.DepthEvent) (domain.DepthEvent, error) {
	return depth, nil
}

func (s *OrderBookRepository) Remove(_ context.Context, _ string) error { return nil }

func (s *OrderBookRepository) Close() error {
	return nil
}
