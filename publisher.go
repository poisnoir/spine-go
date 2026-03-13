package spine

import (
	"io"
	"net"

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
	intEncoder   *mad.Mad[uint32]
	errorEncoder *mad.Mad[string]
	clients      [](client[K])
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

	go p.run()

	return p, nil
}

func (p *Publisher[K]) run() {

}

func (p *Publisher[K]) registerSubscriber(conn io.ReadWriteCloser, data K) {

}

func (p *Publisher[K]) Publish(data K) error {

	packetSize := uint32(p.encoder.GetRequiredSize(&data))
	buf := make([]byte, int(packetSize)+5)

	// publisher is gonna send packets as fast as possible so there is need to distinguish between packets if they combine
	p.intEncoder.Encode(&packetSize, buf)
	buf[4] = globals.PUBLISER_PUSH
	p.encoder.Encode(&data, buf[5:])

	return nil
}

func (p *Publisher[K]) sendToSubscribers(conn io.ReadWriteCloser) {

}

func (p *Publisher[K]) Subscribers() []string {
	return nil
}

type client[K any] struct {
	conn         io.ReadWriteCloser
	dataQueue    chan K
	reconnectSig chan struct{}
}

func newClient[K any](conn io.ReadWriteCloser) *client[K] {
	return &client[K]{
		conn:         conn,
		dataQueue:    make(chan K, 100),
		reconnectSig: make(chan struct{}),
	}
}

func (c *client[K]) run() {
	buf := make([]byte, 1)
	for {
		c.conn.Read(buf)
		if buf[0] != globals.PING_CODE {
			continue
		}
		c.conn.Write([]byte{globals.PONG_CODE})
	}
}
