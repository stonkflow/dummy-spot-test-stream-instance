package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"dummy-spot-test-stream-instance/internal/usecase"
)

const defaultRetryDelay = time.Second

type Service struct {
	name      string
	pipelines []usecase.Pipeline
	closers   []NamedCloser
	logger    *slog.Logger
}

type ServiceOption func(*Service)

func WithServiceName(name string) ServiceOption {
	return func(service *Service) {
		if name != "" {
			service.name = name
		}
	}
}

func NewService(
	pipelines []usecase.Pipeline,
	closers []NamedCloser,
	logger *slog.Logger,
	options ...ServiceOption,
) *Service {
	if logger == nil {
		logger = slog.Default()
	}

	service := &Service{
		name:      "pipeline-service",
		pipelines: pipelines,
		closers:   closers,
		logger:    logger.With("component", "app.service"),
	}

	for _, option := range options {
		if option == nil {
			continue
		}
		option(service)
	}

	return service
}

func (s *Service) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	s.logger.Info("service starting", "service_name", s.name, "pipelines", len(s.pipelines))

	if len(s.pipelines) == 0 {
		return fmt.Errorf("at least one pipeline is required")
	}

	errCh := make(chan error, len(s.pipelines))
	var wg sync.WaitGroup

	for _, p := range s.pipelines {
		pipeline := withPipelineDefaults(p)
		if err := validatePipeline(pipeline); err != nil {
			return fmt.Errorf("invalid pipeline %q: %w", pipeline.Name, err)
		}
		executor := buildPipelineExecutor(pipeline)
		wg.Add(1)
		go func(execute usecase.NextFunc) {
			defer wg.Done()
			if err := s.runPipeline(ctx, pipeline, execute); err != nil {
				select {
				case errCh <- fmt.Errorf("%s: %w", pipeline.Name, err):
				default:
				}
			}
		}(executor)
	}

	var runErr error
	select {
	case <-ctx.Done():
		runErr = ctx.Err()
	case runErr = <-errCh:
		cancel()
	}

	wg.Wait()

	closeErr := s.close()
	if errors.Is(runErr, context.Canceled) {
		runErr = nil
	}
	runErr = errors.Join(runErr, closeErr)
	if runErr == nil {
		return nil
	}
	return runErr
}

func (s *Service) runPipeline(ctx context.Context, pipeline usecase.Pipeline, execute usecase.NextFunc) error {
	for {
		payload, err := pipeline.Source.Receive(ctx)
		if err != nil {
			if isContextDone(err) || isContextDone(ctx.Err()) {
				return nil
			}
			s.logger.Error("pipeline receive failed", "pipeline", pipeline.Name, "error", err)
			if err = sleepWithContext(ctx, pipeline.RetryPolicy.Delay); err != nil {
				return nil
			}
			continue
		}

		if len(payload) == 0 {
			continue
		}

		packet, err := pipeline.Decoder.Decode(ctx, payload)
		if err != nil {
			if isContextDone(err) || isContextDone(ctx.Err()) {
				return nil
			}
			s.logger.Error("pipeline deserialize failed", "pipeline", pipeline.Name, "error", err)
			if err = sleepWithContext(ctx, pipeline.RetryPolicy.Delay); err != nil {
				return nil
			}
			continue
		}

		if packet == nil {
			packet = &usecase.Packet{}
		}
		if len(packet.Raw) == 0 {
			packet.Raw = payload
		}

		if err = execute(ctx, packet); err != nil {
			if errors.Is(err, usecase.ErrSkipPacket) {
				continue
			}
			if isContextDone(err) || isContextDone(ctx.Err()) {
				return nil
			}
			s.logger.Error("pipeline handler failed", "pipeline", pipeline.Name, "error", err)
			if err = sleepWithContext(ctx, pipeline.RetryPolicy.Delay); err != nil {
				return nil
			}
		}
	}
}

func (s *Service) close() error {
	var closeErr error

	for _, closer := range s.closers {
		if closer.Close == nil {
			continue
		}
		if err := closer.Close(); err != nil {
			closeErr = errors.Join(closeErr, fmt.Errorf("close %s: %w", closer.Name, err))
		}
	}

	return closeErr
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

func isContextDone(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func runHandlers(ctx context.Context, handlers []usecase.Handler, packet *usecase.Packet) error {
	for i, handler := range handlers {
		if err := handler.Handle(ctx, packet); err != nil {
			if errors.Is(err, usecase.ErrSkipPacket) {
				return usecase.ErrSkipPacket
			}
			return fmt.Errorf("handler[%d]: %w", i, err)
		}
	}
	return nil
}

func buildPipelineExecutor(pipeline usecase.Pipeline) usecase.NextFunc {
	next := func(ctx context.Context, packet *usecase.Packet) error {
		return runHandlers(ctx, pipeline.Handlers, packet)
	}

	for i := len(pipeline.Middlewares) - 1; i >= 0; i-- {
		middleware := pipeline.Middlewares[i]
		if middleware == nil {
			continue
		}
		next = middleware(next)
	}

	return next
}

func withPipelineDefaults(p usecase.Pipeline) usecase.Pipeline {
	if p.Decoder == nil {
		p.Decoder = usecase.DecoderFunc(usecase.BytesDecoder)
	}
	if p.RetryPolicy.Delay <= 0 {
		p.RetryPolicy.Delay = defaultRetryDelay
	}
	return p
}

func validatePipeline(p usecase.Pipeline) error {
	if p.Name == "" {
		return fmt.Errorf("name is required")
	}
	if p.Source == nil {
		return fmt.Errorf("source is required")
	}
	if p.Decoder == nil {
		return fmt.Errorf("decoder is required")
	}
	if len(p.Handlers) == 0 {
		return fmt.Errorf("at least one handler is required")
	}
	for i, h := range p.Handlers {
		if h == nil {
			return fmt.Errorf("handler at index %d is nil", i)
		}
	}
	for i, middleware := range p.Middlewares {
		if middleware == nil {
			return fmt.Errorf("middleware at index %d is nil", i)
		}
	}
	return nil
}
