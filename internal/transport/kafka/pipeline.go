package kafka

import (
	"context"

	"dummy-spot-test-stream-instance/internal/usecase"
	"dummy-spot-test-stream-instance/internal/usecase/ports"
)

type KeyFunc func(packet *usecase.Packet) []byte

func Source(consumer ports.CommandConsumer) usecase.Source {
	return usecase.SourceFunc(consumer.Receive)
}

func ProduceHandler(producer ports.EventProducer, keyFn KeyFunc, encoder usecase.ValueEncoder) usecase.Handler {
	return usecase.HandlerFunc(func(ctx context.Context, packet *usecase.Packet) error {
		payload, err := usecase.EncodePacket(packet, encoder)
		if err != nil {
			return err
		}

		key := packet.Key
		if keyFn != nil {
			key = keyFn(packet)
		}

		return producer.Send(ctx, key, payload)
	})
}
