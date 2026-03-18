package middleware

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"dummy-spot-test-stream-instance/internal/usecase"
)

func Logging(logger *slog.Logger, pipeline string) usecase.Middleware {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next usecase.NextFunc) usecase.NextFunc {
		return func(ctx context.Context, packet *usecase.Packet) error {
			startedAt := time.Now()
			err := next(ctx, packet)
			elapsed := time.Since(startedAt)

			if err == nil || errors.Is(err, usecase.ErrSkipPacket) {
				logger.Debug("pipeline packet processed", "pipeline", pipeline, "duration", elapsed)
				return err
			}

			logger.Warn("pipeline packet failed", "pipeline", pipeline, "duration", elapsed, "error", err)
			return err
		}
	}
}
