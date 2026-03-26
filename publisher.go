package spine

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"slices"
	"sync"

	"github.com/poisnoir/mad-go"
	"github.com/poisnoir/spine-go/internal/globals"
	"github.com/xtaci/kcp-go/v5"

	"github.com/grandcat/zeroconf"
)

type Publisher[K any] struct {
	namespace    *Namespace
	name         string
	listener     *kcp.Listener
	server       *zeroconf.Server
	encoder      *mad.Mad[K]
	errorEncoder *mad.Mad[string]
	logger       *slog.Logger
	clients      []*client[K]
}

func NewPublisher[K any](ns *Namespace, name string) (*Publisher[K], error) {

	logger := ns.logger.With(
		ns.Name(),
		"publiser",
		name,
		"new publisher",
	)

	encoder, err := mad.NewMad[K]()
	if err != nil {
		return nil, err
	}

	listener, err := kcp.ListenWithOptions(":0", ns.encryption, 10, 3)
	if err != nil {
		logger.Error("unable to create listener", "error", err)
		return nil, err
	}

	server, err := zeroconf.Register(
		name+globals.ZERO_CONF_PUBLISHER_PREFIX,
		globals.ZERO_CONF_TYPE,
		globals.ZERO_CONF_DOMAIN,
		listener.Addr().(*net.UDPAddr).Port,
		[]string{"id=botzilla_publish_" + name},
		nil,
	)

	p := &Publisher[K]{
		namespace: ns,
		name:      name,
		listener:  listener,
		encoder:   encoder,
		server:    server,
	}

	go runListener(listener, logger, p.registerSubscriber)

	return p, nil
}

func (p *Publisher[K]) run() {

}

func (p *Publisher[K]) registerSubscriber(conn io.ReadWriteCloser) {

	var err error
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	bufPtr := p.namespace.bufferPool.Get().(*[]byte)
	defer p.namespace.bufferPool.Put(bufPtr)
	buf := *bufPtr

	_, err = conn.Read(make([]byte, 1))
	if err != nil {
		return
	}

	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	if !slices.Equal([]byte(p.encoder.Code()), buf[:n]) {
		err = fmt.Errorf("invalid data code")
		return
	}

	_, err = conn.Write([]byte{globals.OK_STATUS_CODE})
	if err != nil {
		return
	}

	c := newClient[K](conn)
	p.clients = append(p.clients, c)
}

func (p *Publisher[K]) Publish(data K) error {

	packetSize := uint32(p.encoder.GetRequiredSize(&data))
	buf := make([]byte, int(packetSize)+5)


	return nil
}

func (p *Publisher[K]) sendToSubscribers(conn io.ReadWriteCloser) {
	bufPtr := p.namespace.bufferPool.Get().(*[]byte)
	defer p.namespace.bufferPool.Put(bufPtr)
}

func (p *Publisher[K]) Subscribers() []string {
	return nil
}

type client[K any] struct {
	conn     io.ReadWriteCloser
	data     K
	dataLock sync.RWMutex
	sendSig  chan struct{}
}

func newClient[K any](conn io.ReadWriteCloser) *client[K] {
	c := &client[K]{
		conn:    conn,
		sendSig: make(chan struct{}),
	}

	ping(conn io.ReadWriteCloser)

	return c
}
func (c *client[K]) Close() {}
