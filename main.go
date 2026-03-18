package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"dummy-spot-test-stream-instance/internal/adapters/noop"
	"dummy-spot-test-stream-instance/internal/app"
	codecproto "dummy-spot-test-stream-instance/internal/codec/proto"
	"dummy-spot-test-stream-instance/internal/codec/wsjson"
	"dummy-spot-test-stream-instance/internal/transport/kafka"
	"dummy-spot-test-stream-instance/internal/transport/ws"
	"dummy-spot-test-stream-instance/internal/usecase"
	usecasemw "dummy-spot-test-stream-instance/internal/usecase/middleware"
	"dummy-spot-test-stream-instance/internal/usecase/ports"

	"github.com/alexflint/go-arg"
)

func main() {
	var opts options
	arg.MustParse(&opts)

	if err := validateOptions(opts); err != nil {
		slog.Error("invalid config", "error", err)
		os.Exit(2)
	}
	brokers := splitCSV(opts.Brokers)
	sourceStreams := splitCSV(opts.SourceStreams)

	logger := newLogger(opts.LogLevel)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	consumer := noop.NewCommandConsumer()
	producer := noop.NewEventProducer()
	wsClient := noop.NewWSClient()
	orderBook := noop.NewOrderBookStore()
	metricsCollector := noop.NewMetricsCollector()
	dlqWriter := noop.NewDLQWriter()

	logger.Info("config loaded",
		"ws_url", opts.WSURL,
		"brokers", brokers,
		"topic_command", opts.TopicCommand,
		"topic_event", opts.TopicEvent,
		"source_streams", sourceStreams,
	)

	svc := app.NewService(
		buildPipelines(consumer, producer, wsClient, logger, metricsCollector, dlqWriter),
		buildClosers(consumer, producer, wsClient, orderBook),
		logger,
		app.WithServiceName("dummy-spot-gateway"),
	)

	if err := svc.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("service stopped with error", "error", err)
		os.Exit(1)
	}

	logger.Info("service stopped")
}

type options struct {
	SourceStreams string `arg:"--source-streams,env:SOURCE_STREAMS" placeholder:"string" help:"comma-separated source streams, for example: btcusdt@trade,btcusdt@depth@100ms"`
	WSURL         string `arg:"--ws-url,env:WS_URL,required" placeholder:"string" help:"websocket uri of the market data source"`
	Brokers       string `arg:"--brokers,env:BROKERS,required" placeholder:"string" help:"comma-separated kafka brokers, for example: localhost:9092"`
	TopicCommand  string `arg:"--topic-command,env:TOPIC_COMMAND,required" placeholder:"string" help:"kafka topic for incoming stream subscription management commands"`
	TopicEvent    string `arg:"--topic-event,env:TOPIC_EVENT,required" placeholder:"string" help:"kafka topic for outgoing events"`
	LogLevel      string `arg:"--log-level,env:LOG_LEVEL" default:"info" placeholder:"string" help:"log level: debug, info, warn, error"`
}

func validateOptions(opts options) error {
	var validationErr error

	if opts.WSURL == "" {
		validationErr = errors.Join(validationErr, fmt.Errorf("ws-url is required"))
	}
	if len(splitCSV(opts.Brokers)) == 0 {
		validationErr = errors.Join(validationErr, fmt.Errorf("at least one broker is required"))
	}
	if opts.TopicCommand == "" {
		validationErr = errors.Join(validationErr, fmt.Errorf("topic-command is required"))
	}
	if opts.TopicEvent == "" {
		validationErr = errors.Join(validationErr, fmt.Errorf("topic-event is required"))
	}

	return validationErr
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}

func newLogger(level string) *slog.Logger {
	slogLevel := slog.LevelInfo

	switch strings.ToLower(level) {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	}

	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLevel,
	}))
}

func buildPipelines(
	consumer ports.CommandConsumer,
	producer ports.EventProducer,
	wsClient ports.WSClient,
	logger *slog.Logger,
	metrics usecasemw.MetricsCollector,
	dlq usecasemw.DLQWriter,
) []usecase.Pipeline {
	commandDecoder := codecproto.NewCommandDecoder()
	eventEncoder := codecproto.NewEventEncoder()
	wsCommandEncoder := wsjson.NewCommandEncoder()
	wsEventDecoder := wsjson.NewEventDecoder()

	kafkaToWSName := "kafka-command->ws"
	wsToKafkaName := "ws-event->kafka"

	return []usecase.Pipeline{
		usecase.NewPipeline(
			kafkaToWSName,
			kafka.Source(consumer),
			ws.SendHandler(wsClient, wsCommandEncoder),
		).With(
			usecase.WithDecoder(usecase.DecodeWith(commandDecoder)),
			usecase.AppendMiddlewares(
				usecasemw.Recover(),
				usecasemw.Metrics(metrics, kafkaToWSName),
				usecasemw.Logging(logger, kafkaToWSName),
				usecasemw.DLQ(dlq, kafkaToWSName, logger),
				usecasemw.Retry(usecasemw.RetryConfig{
					Attempts: 3,
					Delay:    250 * time.Millisecond,
				}),
			),
		),
		usecase.NewPipeline(
			wsToKafkaName,
			ws.Source(wsClient),
			kafka.ProduceHandler(producer, nil, eventEncoder),
		).With(
			usecase.WithDecoder(usecase.DecodeWith(wsEventDecoder)),
			usecase.AppendMiddlewares(
				usecasemw.Recover(),
				usecasemw.Metrics(metrics, wsToKafkaName),
				usecasemw.Logging(logger, wsToKafkaName),
				usecasemw.DLQ(dlq, wsToKafkaName, logger),
				usecasemw.Retry(usecasemw.RetryConfig{
					Attempts: 3,
					Delay:    250 * time.Millisecond,
				}),
			),
		),
	}
}

func buildClosers(
	consumer ports.CommandConsumer,
	producer ports.EventProducer,
	wsClient ports.WSClient,
	orderBook ports.OrderBookStore,
) []app.NamedCloser {
	return []app.NamedCloser{
		{Name: "command consumer", Close: consumer.Close},
		{Name: "event producer", Close: producer.Close},
		{Name: "websocket client", Close: wsClient.Close},
		{Name: "order book store", Close: orderBook.Close},
	}
}
