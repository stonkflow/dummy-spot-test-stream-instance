package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"dummy-spot-test-stream-instance/internal/domain"
)

var ErrSkipPacket = errors.New("skip packet")

type Packet struct {
	Raw     []byte
	Payload domain.Payload
	Key     []byte
	Meta    map[string]string
}

type Source interface {
	Receive(ctx context.Context) ([]byte, error)
}

type SourceFunc func(ctx context.Context) ([]byte, error)

func (f SourceFunc) Receive(ctx context.Context) ([]byte, error) {
	return f(ctx)
}

type Decoder interface {
	Decode(ctx context.Context, raw []byte) (*Packet, error)
}

type DecoderFunc func(ctx context.Context, raw []byte) (*Packet, error)

func (f DecoderFunc) Decode(ctx context.Context, raw []byte) (*Packet, error) {
	return f(ctx, raw)
}

type Handler interface {
	Handle(ctx context.Context, packet *Packet) error
}

type HandlerFunc func(ctx context.Context, packet *Packet) error

func (f HandlerFunc) Handle(ctx context.Context, packet *Packet) error {
	return f(ctx, packet)
}

type RetryPolicy struct {
	Delay time.Duration
}

type NextFunc func(ctx context.Context, packet *Packet) error
type Middleware func(next NextFunc) NextFunc

type ValueDecoder interface {
	Decode(raw []byte) (domain.Payload, error)
}

type ValueEncoder interface {
	Encode(payload domain.Payload) ([]byte, error)
}

type PipelineOption func(*Pipeline)

type Pipeline struct {
	Name        string
	Source      Source
	Decoder     Decoder
	Handlers    []Handler
	Middlewares []Middleware
	RetryPolicy RetryPolicy
}

func NewPipeline(name string, source Source, handlers ...Handler) Pipeline {
	return Pipeline{
		Name:     name,
		Source:   source,
		Decoder:  DecoderFunc(BytesDecoder),
		Handlers: handlers,
	}
}

func (p Pipeline) With(options ...PipelineOption) Pipeline {
	for _, option := range options {
		if option == nil {
			continue
		}
		option(&p)
	}
	return p
}

func WithDecoder(decoder Decoder) PipelineOption {
	return func(pipeline *Pipeline) {
		if decoder != nil {
			pipeline.Decoder = decoder
		}
	}
}

func WithRetryPolicy(policy RetryPolicy) PipelineOption {
	return func(pipeline *Pipeline) {
		pipeline.RetryPolicy = policy
	}
}

func WithRetryDelay(delay time.Duration) PipelineOption {
	return func(pipeline *Pipeline) {
		pipeline.RetryPolicy.Delay = delay
	}
}

func AppendHandlers(handlers ...Handler) PipelineOption {
	return func(pipeline *Pipeline) {
		pipeline.Handlers = append(pipeline.Handlers, handlers...)
	}
}

func AppendMiddlewares(middlewares ...Middleware) PipelineOption {
	return func(pipeline *Pipeline) {
		pipeline.Middlewares = append(pipeline.Middlewares, middlewares...)
	}
}

func BytesDecoder(_ context.Context, raw []byte) (*Packet, error) {
	return &Packet{
		Raw: raw,
	}, nil
}

func DecodeWith(decoder ValueDecoder) Decoder {
	return DecoderFunc(func(_ context.Context, raw []byte) (*Packet, error) {
		if decoder == nil {
			return nil, fmt.Errorf("decoder is nil")
		}

		payload, err := decoder.Decode(raw)
		if err != nil {
			return nil, err
		}

		return &Packet{
			Raw:     raw,
			Payload: payload,
		}, nil
	})
}

func EncodePacket(packet *Packet, encoder ValueEncoder) ([]byte, error) {
	if packet == nil {
		return nil, fmt.Errorf("packet is nil")
	}
	if encoder != nil {
		return encoder.Encode(packet.Payload)
	}
	return packet.Bytes()
}

func (packet *Packet) Bytes() ([]byte, error) {
	if packet == nil {
		return nil, fmt.Errorf("packet is nil")
	}
	if !packet.Payload.Empty() {
		return nil, fmt.Errorf("packet has typed payload, encoder is required")
	}
	if len(packet.Raw) > 0 {
		return packet.Raw, nil
	}
	return nil, fmt.Errorf("packet payload is empty")
}
