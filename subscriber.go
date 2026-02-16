package spine

import (
	"context"
)

type Subscriber[K any] struct {
	name         string
	subscribedTo string
	handler      func(K)
	namespace    *Namespace
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewSubscriber[K any](namespace *Namespace, name string, publisherName string, handler func(K)) (*Subscriber[K], error) {

	ctx, cancel := context.WithCancel(namespace.ctx)

	sub := &Subscriber[K]{
		name:         name,
		subscribedTo: publisherName,
		handler:      handler,
		ctx:          ctx,
		cancel:       cancel,
	}

	go sub.start()

	return sub, nil
}

func (s *Subscriber[K]) start() {
	defer s.cancel()

	// waits until ctx expires or errors out in zero conf
	address, err := s.namespace.GetPublisher(s.ctx, s.subscribedTo)
	if err != nil {
		// log the error
		return
	}

	// waits until ctx expires or errors out in connection
	// todo
	// have to implement reconnection logic
	err = s.connect(address)

	if err != nil {
		// log the error
		return
	}
}

func (s *Subscriber[K]) Stop() {
	s.cancel()
}

func (s *Subscriber[K]) SubscribedTo() string {
	return s.subscribedTo
}

func (s *Subscriber[K]) connect(address string) error {

	for {
	}
}
