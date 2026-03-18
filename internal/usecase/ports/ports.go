package ports

import (
	"context"

	"dummy-spot-test-stream-instance/internal/domain"
)

type CommandConsumer interface {
	Receive(ctx context.Context) ([]byte, error)
	Close() error
}

type EventProducer interface {
	Send(ctx context.Context, key []byte, value []byte) error
	Close() error
}

type WSClient interface {
	Send(ctx context.Context, payload []byte) error
	Receive(ctx context.Context) ([]byte, error)
	Close() error
}

type OrderBookRepository interface {
	UpsertDepth(ctx context.Context, depth domain.DepthEvent) (domain.DepthEvent, error)
	Remove(ctx context.Context, symbol string) error
	Close() error
}
