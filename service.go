package spine

import (
	"context"

	"github.com/grandcat/zeroconf"
	"github.com/poisnoir/mad-go"
	"github.com/poisnoir/spine-go/internal/globals"
)

type Service[K any, V any] struct {
	name         string
	server       *zeroconf.Server
	namespace    *Namespace
	context      context.Context
	cancel       context.CancelFunc
	handler      func(K) (V, error)
	keyEncoder   *mad.Mad[K]
	valueEncoder *mad.Mad[V]
	requests     chan K
}

func NewService[K any, V any](namespace *Namespace, name string, handler func(K) (V, error)) (*Service[K, V], error) {

	logger := namespace.logger.With(
		namespace.Name(),
		"service",
		name,
		"service registry",
	)

	ctx, cancel := context.WithCancel(namespace.ctx)

	server, err := zeroconf.Register(
		globals.ZERO_CONF_SERVICE_PREFIX+name,
		namespace.Name()+globals.ZERO_CONF_TYPE,
		globals.ZERO_CONF_DOMAIN,
		0,
		[]string{"id=botzilla_service_" + name},
		nil,
	)

	if err != nil {
		cancel()
		logger.Error("unable to register service to zeroconf", "error", err)
		return nil, err
	}

	s := &Service[K, V]{
		name:      name,
		server:    server,
		namespace: namespace,
		context:   ctx,
		cancel:    cancel,
		requests:  make(chan K, 100),
	}

	go s.handleRequest()
	return s, nil
}

func (s *ServiceCaller[K, V]) handlePackets() {

}

func (s *Service[K, V]) handleRequest() {

	logger := s.namespace.logger.With(
		s.namespace.Name(),
		"service",
		s.name,
		"request handler",
	)

	for {
		select {
		case request := <-s.requests:
			response, err := s.handler(request)
			if err != nil {
				logger.Error("unable to handle request", "error", err)
			}
			s.namespace.logger.Info("handled request", "request", request, "response", response)
		case <-s.context.Done():
			return
		}
	}
}

func (s *Service[K, V]) Close() {
	s.cancel()
}

func (s *Service[K, V]) Close() {
	s.cancel()
}

func (s *Service[K, V]) Name() string {
	return s.name
}
