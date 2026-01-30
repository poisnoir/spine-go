package botzilla

import (
	"context"
	"log/slog"
	"sync"

	"github.com/Pois-Noir/Botzilla/internal/globals"
)

type Namespace struct {
	name            string
	useEncryption   bool   // It is gonna slow down all comm types.
	secretKey       string // key used for generating hmac. Do not share it in the network!!!
	reg             *Registry
	ctx             context.Context
	cancel          context.CancelFunc
	logger          *slog.Logger
	serviceCallPool sync.Pool
}

func JointNamespace(name string, secretKey string, logger *slog.Logger, useEncryption bool) (*Namespace, error) {

	ctx, cancel := context.WithCancel(context.Background())
	ns := &Namespace{
		name:          name,
		useEncryption: useEncryption,
		secretKey:     secretKey,
		cancel:        cancel,
		logger:        logger,
		ctx:           ctx,
		serviceCallPool: sync.Pool{New: func() interface{} {
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

func (ns *Namespace) GetService(ctx context.Context, name string) (string, error) {
	return ns.reg.Lookup(ctx, globals.ZERO_CONF_SERVICE_PREFIX+name)
}

func (ns *Namespace) GetPublisher(ctx context.Context, name string) (string, error) {
	return ns.reg.Lookup(ctx, globals.ZERO_CONF_PUBLISHER_PREFIX+name)
}
