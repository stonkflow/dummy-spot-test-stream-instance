package ports

import "context"

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

type OrderBookStore interface {
	Remove(symbol string)
	Close() error
}
