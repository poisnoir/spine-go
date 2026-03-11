package spine

import (
	"context"
	"fmt"
	"time"

	"github.com/poisnoir/mad-go"
	"github.com/poisnoir/spine-go/internal/globals"
	"github.com/xtaci/kcp-go/v5"
)

type Subscriber[K any] struct {
	subscribedTo string
	handler      func(K)
	conn         *kcp.UDPSession
	namespace    *Namespace
	ctx          context.Context
	cancel       context.CancelFunc
	decoder      *mad.Mad[K]
	isConnected  bool
}

func NewSubscriber[K any](namespace *Namespace, topic string, handler func(K)) (*Subscriber[K], error) {

	decoder, err := mad.NewMad[K]()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(namespace.ctx)

	sub := &Subscriber[K]{
		subscribedTo: topic,
		handler:      handler,
		ctx:          ctx,
		cancel:       cancel,
		decoder:      decoder,
		isConnected:  false,
	}

	go sub.run()
	go sub.runHandler()

	return sub, nil
}

func (s *Subscriber[K]) run() {
	defer s.cancel()

	// publisher should send ping packet every 10 seconds
	// or send data. If it doesn't happen we assume connection
	// is dead and try to reconnect
	// note: 5 extra seconds is give for error in network
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	if s.isConnected {
		select {
		case <-ticker.C:
			// maybe move it to publisher
			s.isConnected = false
		case <-s.ctx.Done():
			return
		}
	} else {
		s.connect()
	}

}

func (s *Subscriber[K]) runHandler() {

	_ = s.namespace.logger.With(
		s.namespace.Name(),
		"subscriber",
		s.subscribedTo,
		"run handler",
	)

	bufPtr := s.namespace.bufferPool.Get().(*[]byte)
	defer s.namespace.bufferPool.Put(bufPtr)
	buf := *bufPtr

	for {
		n, err := s.conn.Read(buf)
		// Todo
		if err != nil {

		}

		if buf[0] == globals.PING_CODE && n == 1 {
			continue
		}

		if buf[0] != globals.PUBLISER_PUSH {
			continue
		}

	}
}

func (s *Subscriber[K]) connect() error {

	logger := s.namespace.logger.With(
		s.namespace.Name(),
		"subscriber",
		s.subscribedTo,
		"connect",
	)

	// finding the service
	address, err := s.namespace.GetService(s.subscribedTo, s.ctx)
	if err != nil {
		logger.Error("unable to find the service", "error", err)
		return err // the only way to fail here is to run out of context
	}

	// establishing connection
	sess, err := kcp.DialWithOptions(address, s.namespace.encryption, 10, 3)
	if err != nil {
		logger.Error("failed to dial service", "error", err)
		return err
	}

	// getting buffer for comm
	bufPtr := s.namespace.bufferPool.Get().(*[]byte)
	defer s.namespace.bufferPool.Put(bufPtr)
	buf := *bufPtr

	// validating input/output service types
	keyCode := s.decoder.Code()
	n := copy(buf, keyCode)

	n, err = write(sess, buf, n, true)
	if err != nil {
		logger.Error("failed to validate service input type", "error", err)
		return err
	} else if n != 1 {
		err = fmt.Errorf("response is corrupted")
		logger.Error("failed to validate service input type", "error", err)
		return err
	} else if buf[0] != globals.OK_STATUS_CODE {
		err = fmt.Errorf("service data type is different")
		logger.Error("failed to validate service input type", "error", err)
		return err
	}

	s.conn = sess
	s.isConnected = true

	return nil
}

func (s *Subscriber[K]) Stop() {
	s.cancel()
}

func (s *Subscriber[K]) SubscribedTo() string {
	return s.subscribedTo
}
