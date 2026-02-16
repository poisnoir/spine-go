package spine

import (
	"context"
	"net"

	"github.com/poisnoir/spine-go/internal/globals"

	"github.com/grandcat/zeroconf"
)

type Publisher[K any] struct {
	name     string
	listener *net.UDPConn // used for incoming messages
	server   *zeroconf.Server
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

}

func (p *Publisher[K]) registerSubscriber(conn net.Conn) {

}

func (p *Publisher[K]) Publish(data K) error {

	return nil
}

func (p *Publisher[K]) Subscribers() []string {
	return nil
}

type client[K any] struct {
	Path       chan K
	connection net.Conn
	ctx        context.Context
	cancel     context.CancelFunc
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

}
