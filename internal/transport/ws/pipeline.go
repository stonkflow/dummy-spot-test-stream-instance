package ws

import (
	"context"

	"dummy-spot-test-stream-instance/internal/usecase"
	"dummy-spot-test-stream-instance/internal/usecase/ports"
)

func Source(client ports.WSClient) usecase.Source {
	return usecase.SourceFunc(client.Receive)
}

func SendHandler(client ports.WSClient, encoder usecase.ValueEncoder) usecase.Handler {
	return usecase.HandlerFunc(func(ctx context.Context, packet *usecase.Packet) error {
		payload, err := usecase.EncodePacket(packet, encoder)
		if err != nil {
			return err
		}
		return client.Send(ctx, payload)
	})
}
