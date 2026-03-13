package spine

import (
	"context"
	"fmt"
	"io"

	"github.com/grandcat/zeroconf"
	"github.com/poisnoir/mad-go"
	"github.com/xtaci/kcp-go/v5"
)

type ThreadedService[K any, V any] struct {
	name         string
	server       *zeroconf.Server
	namespace    *Namespace
	context      context.Context
	listener     *kcp.Listener
	cancel       context.CancelFunc
	handler      func(K) (V, error)
	keyEncoder   *mad.Mad[K]
	valueEncoder *mad.Mad[V]
	errorEncoder *mad.Mad[string]
	requests     chan serviceRequest[K, V]
}

func NewThreadedService[K any, V any](namespace *Namespace, name string, handler func(K) (V, error)) (*ThreadedService[K, V], error) {

	keyEnc, valueEnc, listener, server, err := generateService[K, V](namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %v", err)
	}

	errEnc, _ := mad.NewMad[string]()
	ctx, cancel := context.WithCancel(namespace.ctx)

	ts := &ThreadedService[K, V]{
		name:         name,
		server:       server,
		namespace:    namespace,
		handler:      handler,
		context:      ctx,
		listener:     listener,
		keyEncoder:   keyEnc,
		valueEncoder: valueEnc,
		errorEncoder: errEnc,
		cancel:       cancel,
		requests:     make(chan serviceRequest[K, V], 100),
	}

	// todo fix me pls
	logger := namespace.logger

	go runListener(listener, logger, ts.clientHandler) // stops when listener closes
	return ts, nil
}

func (s *ThreadedService[K, V]) clientHandler(conn io.ReadWriteCloser) {

	logger := s.namespace.logger.With(
		s.namespace.Name(),
		"service",
		s.name,
		"client handler",
	)

	bufPtr := s.namespace.bufferPool.Get().(*[]byte)
	defer s.namespace.bufferPool.Put(bufPtr)

	handleClient(conn, s.keyEncoder, s.valueEncoder, *bufPtr, s.processRequest, logger)

}

func (ts *ThreadedService[K, V]) processRequest(key K) serviceOutput[V] {
	// send to handler

	result, err := ts.handler(key)
	return serviceOutput[V]{data: result, err: err}
}
