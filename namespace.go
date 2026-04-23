package spine

import (
	"context"
	"crypto/sha256"
	"log/slog"
	"sync"

	"github.com/grandcat/zeroconf"
	"github.com/poisnoir/mad-go"
	"github.com/xtaci/kcp-go/v5"
)

type Namespace struct {
	name string
	reg  *Registry

	ctx    context.Context
	cancel context.CancelFunc

	encryption kcp.BlockCrypt // will be set to nill if useEncryption is false, will use hmac instead

	logger           *slog.Logger
	bufferPool       sync.Pool
	stringSerializer *mad.Mad[string]

	listener *kcp.Listener
	server   *zeroconf.Server
}

func JointNamespace(name string, secretKey string, logger *slog.Logger) (*Namespace, error) {

	key := sha256.Sum256([]byte(secretKey))
	encryption, err := kcp.NewAESBlockCrypt(key[:])
	if err != nil {
		return nil, err
	}

	stringSer, _ := mad.NewMad[string]()

	ctx, cancel := context.WithCancel(context.Background())
	ns := &Namespace{
		name: name,

		encryption: encryption,

		ctx:    ctx,
		cancel: cancel,

		logger: logger,
		bufferPool: sync.Pool{New: func() any {
			b := make([]byte, 4096)
			return &b
		}},
		stringSerializer: stringSer,
	}
	reg, err := NewRegistry(ns)
	if err != nil {
		cancel()
		return nil, err
	}
	ns.reg = reg
	return ns, nil
}

func (ns *Namespace) Disconnect() {
	ns.cancel()
}

func (ns *Namespace) Name() string {
	return ns.name
}

func (ns *Namespace) Logger() *slog.Logger {
	return ns.logger
}

func (ns *Namespace) Context() context.Context {
	return ns.ctx
}

func (ns *Namespace) GetService(name string, ctx context.Context) (string, error) {
	return ns.reg.Lookup(ctx, name)
}

func (ns *Namespace) GetPublisher(name string, ctx context.Context) (string, error) {
	return ns.reg.Lookup(ctx, name)
}
