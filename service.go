package botzilla

import (
	"context"
	"time"

	"github.com/Pois-Noir/Botzilla/internal/globals"
	"github.com/Pois-Noir/Botzilla/internal/tcp"
	"github.com/grandcat/zeroconf"

	"github.com/vmihailenco/msgpack/v5"
)

type Service[K any, V any] struct {
	name     string
	server   *zeroconf.Server
	listener *tcp.Listener[K, V]
}

func NewService[K any, V any](name string, maxHandler uint32, handler func(K) (V, error)) (*Service[K, V], error) {

	logger := GetLogger().With(
		"service", name,
		"component", "registry",
	)

	listener, err := tcp.NewListener(handler, maxHandler, logger)
	if err != nil {
		logger.Error("failed to create listener", "error", err)
		return nil, err
	}

	server, err := zeroconf.Register(
		name+globals.ZERO_CONF_SERVICE_PREFIX,
		globals.ZERO_CONF_SERVICE,
		globals.ZERO_CONF_DOMAIN,
		listener.Port(),
		[]string{"id=botzilla_service_" + name},
		nil,
	)

	if err != nil {
		logger.Error("unable to register service to zeroconf", "error", err)
		return nil, err
	}

	go listener.Start()

	return &Service[K, V]{
		name:     name,
		server:   server,
		listener: listener,
	}, nil
}

func (s *Service[K, V]) Name() string {
	return s.name
}

func Call[K any, V any](ctx context.Context, serviceName string, payload K) (V, error) {

	logger := GetLogger().With(
		"operation", "rpc_call",
		"target_service", serviceName,
	)

	var response V

	serviceIP, err := GetService(ctx, serviceName)
	if err != nil {
		logger.ErrorContext(ctx, "service discovery failed", "error", err)
		return response, err
	}

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

	err = msgpack.Unmarshal(encodedResponsePayload, &response)
	if err != nil {
		logger.ErrorContext(ctx, "response unmarshal failed", "error", err)
	}

	logger.DebugContext(ctx, "service call successful", "duration_ms", duration.Milliseconds())
	return response, err
}
