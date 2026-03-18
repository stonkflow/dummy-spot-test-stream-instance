package middleware

import (
	"context"
	"fmt"

	"dummy-spot-test-stream-instance/internal/usecase"
)

func Recover() usecase.Middleware {
	return func(next usecase.NextFunc) usecase.NextFunc {
		return func(ctx context.Context, packet *usecase.Packet) (err error) {
			defer func() {
				if recovered := recover(); recovered != nil {
					err = fmt.Errorf("panic recovered: %v", recovered)
				}
			}()
			return next(ctx, packet)
		}
	}
}
