package spine

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"slices"
	"sync"
	"time"

	"github.com/poisnoir/mad-go"
	"github.com/poisnoir/spine-go/internal/globals"
	"github.com/xtaci/kcp-go/v5"

	"github.com/grandcat/zeroconf"
)

type Publisher[K any] struct {
	name      string
	namespace *Namespace
	server    *zeroconf.Server
	logger    *slog.Logger
	encoder   *mad.Mad[K]

	listener   *kcp.Listener
	clients    []io.ReadWriteCloser
	clientMu   sync.RWMutex
	deadClient chan io.ReadWriteCloser

	sendSig    chan struct{}
	lastDataMu sync.RWMutex
	lastData   K
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
		globals.ZERO_CONF_SERVICE_PREFIX+name,
		ns.Name()+globals.ZERO_CONF_TYPE,
		globals.ZERO_CONF_DOMAIN,
		listener.Addr().(*net.UDPAddr).Port,
		[]string{"id=spine_service_" + name},
		nil,
	)

	p := &Publisher[K]{
		namespace:  ns,
		name:       name,
		listener:   listener,
		encoder:    encoder,
		server:     server,
		logger:     logger,
		sendSig:    make(chan struct{}, 1),
		deadClient: make(chan io.ReadWriteCloser, 64),
		clients:    make([]io.ReadWriteCloser, 0),
	}

	go runListener(listener, logger, p.registerSubscriber)
	go p.run()

	return p, nil
}

func (p *Publisher[K]) run() {

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.sendSig:
			p.lastDataMu.RLock()
			tempData := p.lastData
			p.lastDataMu.RUnlock()

			payloadSize := p.encoder.GetRequiredSize(&tempData)
			if payloadSize > globals.MAX_PACKET_SIZE {
				p.logger.Error("payload size too big", "size", payloadSize)
				continue
			}
			ticker.Reset(10 * time.Second)
			bufPtr := p.namespace.bufferPool.Get().(*[]byte)
			buf := *bufPtr
			p.encoder.Encode(&tempData, buf)

			var wg sync.WaitGroup

			p.clientMu.RLock()
			snapClients := make([]io.ReadWriteCloser, len(p.clients))
			copy(snapClients, p.clients)
			p.clientMu.RUnlock()

			for _, client := range snapClients {
				wg.Add(1)
				go func(target io.ReadWriteCloser) {
					_, err := write(target, buf, payloadSize, false)
					if err != nil {
						select {
						case p.deadClient <- target:
						default:
						}
					}
					wg.Done()
				}(client)
			}

			go func(b *[]byte) {
				wg.Wait()
				p.namespace.bufferPool.Put(b)
			}(bufPtr)

		case deadClient := <-p.deadClient:
			p.clientMu.Lock()
			p.clients = slices.DeleteFunc(p.clients, func(c io.ReadWriteCloser) bool {
				return c == deadClient
			})
			deadClient.Close()
			p.clientMu.Unlock()

		case <-ticker.C:
			p.clientMu.RLock()
			snapClients := make([]io.ReadWriteCloser, len(p.clients))
			copy(snapClients, p.clients)
			p.clientMu.RUnlock()
			for _, client := range snapClients {
				go func(conn io.ReadWriteCloser) {
					err := ping(conn)
					if err != nil {
						select {
						case p.deadClient <- conn:
						default:
						}
					}
				}(client)
			}
		}
	}
}

func (p *Publisher[K]) registerSubscriber(conn io.ReadWriteCloser) {

	var err error
	bufPtr := p.namespace.bufferPool.Get().(*[]byte)
	buf := *bufPtr

	defer func() {
		if err != nil {
			conn.Close()
		}
		p.namespace.bufferPool.Put(bufPtr)
	}()

	n, err := conn.Read(buf)
	if err != nil {
		return
	}

	if !slices.Equal([]byte(p.encoder.Code()), buf[:n]) {
		err = fmt.Errorf("invalid data code")
		conn.Write([]byte{globals.ERROR_MISMATCH_PAYLOAD_CODE})
		return
	}

	_, err = conn.Write([]byte{globals.OK_STATUS_CODE})
	if err != nil {
		return
	}

	p.clientMu.Lock()
	p.clients = append(p.clients, conn)
	p.clientMu.Unlock()

}

func (p *Publisher[K]) Publish(data K) {
	p.lastDataMu.Lock()
	p.lastData = data
	p.lastDataMu.Unlock()

	select {
	case p.sendSig <- struct{}{}:
	default:
	}
}
