package botzilla

import (
	"context"
	"errors"
	"net"

	"github.com/Pois-Noir/Botzilla/internal/globals"
	"github.com/Pois-Noir/Botzilla/internal/utils"

	"github.com/grandcat/zeroconf"
	"github.com/vmihailenco/msgpack/v5"
)

type Publisher[K any] struct {
	name        string
	subscribers *utils.SafeArr[*client[K]] // multiple threads are accessing the arr
	listener    *net.UDPConn               // used for incoming messages
	server      *zeroconf.Server
}

func NewPublisher[K any](name string) (*Publisher[K], error) {

	addr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}

	server, err := zeroconf.Register(
		name+globals.ZERO_CONF_PUBLISHER_PREFIX,
		globals.ZERO_CONF_TYPE,
		globals.ZERO_CONF_DOMAIN,
		addr.Port,
		[]string{"id=botzilla_publish_" + name},
		nil,
	)

	p := &Publisher[K]{
		name:     name,
		listener: conn,
		server:   server,
	}

	go p.startListener()

	return p, nil
}

func (p *Publisher[K]) startListener() {
	for {
		conn, err := p.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				break
			}
			continue
		}
		go p.registerSubscriber(conn)
	}
}

func (p *Publisher[K]) registerSubscriber(conn net.Conn) {
	newC := newClient[K](conn)
	go newC.start()
	p.subscribers.Add(newC)
}

func (p *Publisher[K]) Publish(data K) error {

	allSubs := p.subscribers.Get()

	// todo
	// bottleneck drowning fix
	// should add a timeout mechanism
	for _, subscriber := range allSubs {
		subscriber.Path <- data
	}

	return nil
}

func (p *Publisher[K]) Subscribers() []string {
	return nil
}

type client[K any] struct {
	Path                chan K
	connection          net.Conn
	consecutiveFailures *utils.SafeCounter // if it reaches 3 I probably should remove it :O or ask for a new connection
	ctx                 context.Context
	cancel              context.CancelFunc
}

func newClient[K any](conn net.Conn) *client[K] {
	c := &client[K]{
		Path:       make(chan K),
		ctx:        context.Background(),
		connection: conn,
	}

	return c
}

func (c *client[K]) start() {
	for {
		data := <-c.Path
		message, err := msgpack.Marshal(data)
		// Todo
		// Add log
		if err != nil {
			continue
		}

		_, err = c.connection.Write(message)

	}
}
