package noop

import (
	"context"
	"time"

	"dummy-spot-test-stream-instance/internal/usecase/middleware"
)

type MetricsCollector struct{}

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}

func (collector *MetricsCollector) Increment(_ string, _ map[string]string) {}

func (collector *MetricsCollector) ObserveDuration(_ string, _ time.Duration, _ map[string]string) {}

type DLQWriter struct{}

func NewDLQWriter() *DLQWriter {
	return &DLQWriter{}
}

func (writer *DLQWriter) Write(_ context.Context, _ middleware.DLQEntry) error {
	return nil
}
