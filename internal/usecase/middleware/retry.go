package middleware

import (
	"context"
	"errors"
	"time"

	"dummy-spot-test-stream-instance/internal/usecase"
)

type RetryConfig struct {
	Attempts int
	Delay    time.Duration
}

func Retry(config RetryConfig) usecase.Middleware {
	attempts := config.Attempts
	if attempts <= 0 {
		attempts = 1
	}

	delay := config.Delay
	if delay <= 0 {
		delay = 200 * time.Millisecond
	}

	return func(next usecase.NextFunc) usecase.NextFunc {
		return func(ctx context.Context, packet *usecase.Packet) error {
			var err error

			for attempt := 1; attempt <= attempts; attempt++ {
				err = next(ctx, packet)
				if err == nil || errors.Is(err, usecase.ErrSkipPacket) {
					return err
				}
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return err
				}
				if attempt == attempts {
					return err
				}
				if sleepErr := sleepWithContext(ctx, delay); sleepErr != nil {
					return sleepErr
				}
			}

			return err
		}
	}
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
