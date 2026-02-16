package spine

import (
	"context"
	"log/slog"
	"sync"

	"github.com/poisnoir/spine-go/internal/globals"
	"github.com/xtaci/kcp-go/v5"
)

type Namespace struct {
	name       string
	secretKey  string // key used for generating hmac. Do not share it in the network!!!
	reg        *Registry
	ctx        context.Context
	cancel     context.CancelFunc
	logger     *slog.Logger
	bufferPool sync.Pool
	enyption   kcp.BlockCrypt // will be set to nill if useEncryption is false, will use hmac instead
}

func JointNamespace(name string, secretKey string, logger *slog.Logger, useEncryption bool) (*Namespace, error) {

	var enyption kcp.BlockCrypt = nil
	if useEncryption {
		var err error
		enyption, err = kcp.NewAESBlockCrypt([]byte(secretKey))
		if err != nil {
			return nil, err
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	ns := &Namespace{
		name:      name,
		secretKey: secretKey,
		enyption:  enyption,
		cancel:    cancel,
		logger:    logger,
		ctx:       ctx,
		bufferPool: sync.Pool{New: func() interface{} {
			b := make([]byte, 4096)
			return &b
		}},
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
	return ns.reg.Lookup(ctx, globals.ZERO_CONF_SERVICE_PREFIX+name)
}

func (ns *Namespace) GetPublisher(ctx context.Context, name string) (string, error) {
	return ns.reg.Lookup(ctx, globals.ZERO_CONF_PUBLISHER_PREFIX+name)
}
