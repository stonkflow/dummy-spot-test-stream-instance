package middleware

import (
	"context"
	"errors"
	"time"

	"dummy-spot-test-stream-instance/internal/usecase"
)

type MetricsCollector interface {
	Increment(name string, labels map[string]string)
	ObserveDuration(name string, duration time.Duration, labels map[string]string)
}

func Metrics(collector MetricsCollector, pipeline string) usecase.Middleware {
	return func(next usecase.NextFunc) usecase.NextFunc {
		return func(ctx context.Context, packet *usecase.Packet) error {
			if collector == nil {
				return next(ctx, packet)
			}

			labels := map[string]string{
				"pipeline": pipeline,
			}

			startedAt := time.Now()
			collector.Increment("pipeline_packets_total", labels)

			err := next(ctx, packet)
			collector.ObserveDuration("pipeline_packet_duration", time.Since(startedAt), labels)

			if err == nil || errors.Is(err, usecase.ErrSkipPacket) {
				collector.Increment("pipeline_packets_success_total", labels)
				return err
			}

			collector.Increment("pipeline_packets_failed_total", labels)
			return err
		}
	}
}
