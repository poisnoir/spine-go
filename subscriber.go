package botzilla

import (
	"context"
	"io"
	"net"

	"github.com/Pois-Noir/Botzilla/internal/globals"
	"github.com/Pois-Noir/Botzilla/internal/tcp"
	"github.com/Pois-Noir/Botzilla/internal/utils"
	"github.com/vmihailenco/msgpack/v5"
)

type Subscriber[K any] struct {
	name         string
	subscribedTo string
	handler      func(K)
	isConnected  *utils.SafeVar[bool]
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewSubscriber[K any](name string, publisherName string, handler func(K)) (*Subscriber[K], error) {

	ctx, cancel := context.WithCancel(context.Background())

	sub := &Subscriber[K]{
		name:         name,
		subscribedTo: publisherName,
		handler:      handler,
		isConnected:  utils.NewSafeVar(false),
		ctx:          ctx,
		cancel:       cancel,
	}

	go sub.start()

	return sub, nil
}

func (s *Subscriber[K]) start() {
	defer s.cancel()

	// waits until ctx expires or errors out in zero conf
	address, err := GetPublisher(s.ctx, s.subscribedTo)
	if err != nil {
		// log the error
		return
	}
	s.isConnected.Set(true)

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

func (s *Subscriber[K]) IsConnected() bool {
	return s.isConnected.Get()
}

func (s *Subscriber[K]) connect(address string) error {
	d := net.Dialer{}
	conn, err := d.DialContext(s.ctx, "tcp", address)
	if err != nil {
		return err
	}

	defer conn.Close()

	var headerBuf [globals.HEADER_LENGTH]byte

	for {
		if _, err := io.ReadFull(conn, headerBuf[:]); err != nil {
			return err
		}

		header, err := tcp.Decode(headerBuf[:])
		if err != nil {
			return err
		}

		// Read payload
		payload := make([]byte, header.PayloadLength)
		if _, err := io.ReadFull(conn, payload); err != nil {
			return err
		}

		// Unmarshal and call handler
		var data K
		if err := msgpack.Unmarshal(payload, &data); err != nil {
			continue // Skip malformed messages
		}

		// Call user handler
		s.handler(data)
	}
}
