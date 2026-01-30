package botzilla

import (
	"bytes"
	"context"
	"time"

	"github.com/Pois-Noir/Botzilla/internal/globals"
	"github.com/Pois-Noir/Botzilla/internal/tcp"
	"github.com/grandcat/zeroconf"

	"github.com/vmihailenco/msgpack/v5"
)

type Service[K any, V any] struct {
	name      string
	server    *zeroconf.Server
	listener  *tcp.Listener[K, V]
	namespace *Namespace
	context   context.Context
	cancel    context.CancelFunc
}

func NewService[K any, V any](namespace *Namespace, name string, handler func(K) (V, error)) (*Service[K, V], error) {

	logger := namespace.logger.With(
		namespace.Name(),
		"service",
		name,
	)

	ctx, cancel := context.WithCancel(namespace.ctx)

	listener, err := tcp.NewListener(handler, 1, logger, ctx)
	if err != nil {
		cancel()
		logger.Error("failed to create listener", "error", err)
		return nil, err
	}

	server, err := zeroconf.Register(
		globals.ZERO_CONF_SERVICE_PREFIX+name,
		namespace.Name()+globals.ZERO_CONF_TYPE,
		globals.ZERO_CONF_DOMAIN,
		listener.Port(),
		[]string{"id=botzilla_service_" + name},
		nil,
	)

	if err != nil {
		cancel()
		logger.Error("unable to register service to zeroconf", "error", err)
		return nil, err
	}

	go listener.Start()

	return &Service[K, V]{
		name:      name,
		server:    server,
		listener:  listener,
		namespace: namespace,
		context:   ctx,
		cancel:    cancel,
	}, nil
}

func (s *Service[K, V]) Close() {
	s.cancel()
}

func (s *Service[K, V]) Name() string {
	return s.name
}

func Call[K any, V any](namespace *Namespace, ctx context.Context, serviceName string, payload K) (result V, failed error) {

	logger := namespace.logger.With(
		namespace.Name(),
		"service_call",
		"target_service",
		serviceName,
	)

	var response V

	serviceIP, err := namespace.GetService(ctx, serviceName)
	if err != nil {
		logger.ErrorContext(ctx, "service discovery failed", "error", err)
		return response, err
	}

	defer func() {
		// somehow service failed
		// we will update catch to rediscover the service
		if err != nil {
			namespace.reg.RemoveOnFailure(serviceName)
		}
	}()

	encodedRequestPayload, err := msgpack.Marshal(payload)
	if err != nil {
		logger.ErrorContext(ctx, "request marshal failed", "error", err)
		return response, err
	}

	start := time.Now()
	encodedResponsePayload, err := tcp.Request(ctx, serviceIP, encodedRequestPayload)
	duration := time.Since(start)
	if err != nil {
		logger.ErrorContext(ctx, "network request failed",
			"ip", serviceIP,
			"duration_ms", duration.Milliseconds(),
			"error", err,
		)
		return response, err
	}

	buf := bytes.NewBuffer(encodedResponsePayload)
	dec := msgpack.NewDecoder(buf)

	dec.SetCustomStructTag("msgpack")
	err = dec.Decode(&response)
	if err != nil {
		logger.ErrorContext(ctx, "response unmarshal failed", "error", err)
	}

	logger.DebugContext(ctx, "service call successful", "duration_ms", duration.Milliseconds())
	return response, err
}
