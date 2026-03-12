package spine

import (
	"net"

	"github.com/poisnoir/mad-go"
	"github.com/poisnoir/spine-go/internal/globals"
	"github.com/xtaci/kcp-go/v5"

	"github.com/grandcat/zeroconf"
)

type Publisher[K any] struct {
	namespace *Namespace
	name      string
	listener  *kcp.Listener
	server    *zeroconf.Server
	encoder   *mad.Mad[K]
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

	for {

	}

}

func (p *Publisher[K]) registerSubscriber(conn net.Conn) {

}

func (p *Publisher[K]) Publish(data K) error {

	return nil
}

func (p *Publisher[K]) Subscribers() []string {
	return nil
}
