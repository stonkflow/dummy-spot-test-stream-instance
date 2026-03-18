package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"dummy-spot-test-stream-instance/internal/usecase"
)

type DLQEntry struct {
	Pipeline string
	Key      []byte
	Payload  []byte
	Error    string
	FailedAt time.Time
}

type DLQWriter interface {
	Write(ctx context.Context, entry DLQEntry) error
}

func DLQ(writer DLQWriter, pipeline string, logger *slog.Logger) usecase.Middleware {
	return func(next usecase.NextFunc) usecase.NextFunc {
		return func(ctx context.Context, packet *usecase.Packet) error {
			err := next(ctx, packet)
			if err == nil || errors.Is(err, usecase.ErrSkipPacket) {
				return err
			}
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return err
			}
			if writer == nil {
				return err
			}

			payload := append([]byte(nil), packet.Raw...)
			if len(payload) == 0 {
				encoded, payloadErr := packet.Bytes()
				if payloadErr == nil {
					payload = encoded
				}
			}
			writeErr := writer.Write(ctx, DLQEntry{
				Pipeline: pipeline,
				Key:      append([]byte(nil), packet.Key...),
				Payload:  append([]byte(nil), payload...),
				Error:    err.Error(),
				FailedAt: time.Now().UTC(),
			})
			if writeErr != nil {
				return errors.Join(err, fmt.Errorf("dlq write failed: %w", writeErr))
			}

			if logger != nil {
				logger.Error("packet sent to dlq", "pipeline", pipeline, "error", err)
			}
			return usecase.ErrSkipPacket
		}
	}
}
