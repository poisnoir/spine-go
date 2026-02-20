package spine

import (
	"context"
	"io"
	"net"

	"github.com/grandcat/zeroconf"
	"github.com/poisnoir/mad-go"
	"github.com/poisnoir/spine-go/internal/globals"
	"github.com/xtaci/kcp-go/v5"
)

type Service[K any, V any] struct {
	name         string
	server       *zeroconf.Server
	namespace    *Namespace
	context      context.Context
	listener     *kcp.Listener
	cancel       context.CancelFunc
	handler      func(K) (V, error)
	keyEncoder   *mad.Mad[K]
	valueEncoder *mad.Mad[V]
	requests     chan handlerRequest[K, V]
}

type handlerRequest[K any, V any] struct {
	key    K
	Result chan V
}

func NewService[K any, V any](namespace *Namespace, name string, handler func(K) (V, error)) (*Service[K, V], error) {

	logger := namespace.logger.With(
		namespace.Name(),
		"service",
		name,
		"new service",
	)

	keyEnc, err := mad.NewMad[K]()
	if err != nil {
		logger.Error("unable to create key encoder", "error", err)
		return nil, err
	}

	valueEnc, err := mad.NewMad[V]()
	if err != nil {
		logger.Error("unable to create value encoder", "error", err)
		return nil, err
	}

	listener, err := kcp.ListenWithOptions(":0", namespace.encryption, 10, 3)
	if err != nil {
		logger.Error("unable to create listener", "error", err)
		return nil, err
	}

	ctx, cancel := context.WithCancel(namespace.ctx)

	server, err := zeroconf.Register(
		globals.ZERO_CONF_SERVICE_PREFIX+name,
		namespace.Name()+globals.ZERO_CONF_TYPE,
		globals.ZERO_CONF_DOMAIN,
		listener.Addr().(*net.UDPAddr).Port,
		[]string{"id=spine_service_" + name},
		nil,
	)

	if err != nil {
		cancel()
		logger.Error("unable to register service to zeroconf", "error", err)
		return nil, err
	}

	s := &Service[K, V]{
		name:         name,
		server:       server,
		namespace:    namespace,
		context:      ctx,
		listener:     listener,
		keyEncoder:   keyEnc,
		valueEncoder: valueEnc,
		cancel:       cancel,
		requests:     make(chan handlerRequest[K, V], 100),
	}

	go s.runListener()
	go s.runHandler()
	return s, nil
}

func (s *Service[K, V]) runListener() {

	logger := s.namespace.logger.With(
		s.namespace.Name(),
		"service",
		s.name,
		"run listener",
	)

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			logger.Error("unable to accept connection", "error", err)
			continue
		}

		go s.handleClient(conn)
	}

}

func (s *Service[K, V]) handleClient(conn io.ReadWriteCloser) {
	logger := s.namespace.logger.With(
		s.namespace.Name(),
		"service",
		s.name,
		"client handler",
	)

	bufPtr := s.namespace.bufferPool.Get().(*[]byte)
	defer s.namespace.bufferPool.Put(bufPtr)
	n, err := conn.Read(*bufPtr)
	if err != nil {
		logger.Error("unable to read from connection", "error", err)
		conn.Close()
		return
	}
	var key K
	err = s.keyEncoder.Decode((*bufPtr)[:n], &key)
	if err != nil {
		logger.Error("unable to decode key", "error", err)
		conn.Close()
		return
	}

	result := make(chan V, 1)
	hr := handlerRequest[K, V]{
		key:    key,
		Result: result,
	}

	s.requests <- hr
	res := <-result
	// reseting buffer
	(*bufPtr) = (*bufPtr)[:0]
	s.valueEncoder.Encode(&res, *bufPtr)
	conn.Write((*bufPtr)[:s.valueEncoder.GetRequiredSize(&res)])

	defer conn.Close()
}

func (s *Service[K, V]) runHandler() {

	logger := s.namespace.logger.With(
		s.namespace.Name(),
		"service",
		s.name,
		"request handler",
	)

	for {
		select {
		case request := <-s.requests:
			response, err := s.handler(request.key)
			if err != nil {
				logger.Error("unable to handle request", "error", err)
				// Todo: Change To return error instead of default
			}
			request.Result <- response
			s.namespace.logger.Info("handled request", "request", request, "response", response)
		case <-s.context.Done():
			return
		}
	}
}

func (s *Service[K, V]) Close() {
	s.cancel()
}

func (s *Service[K, V]) Name() string {
	return s.name
}
