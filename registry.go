package spine

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"strconv"
	"sync"

	"github.com/grandcat/zeroconf"
	"github.com/poisnoir/spine-go/internal/globals"
)

// Todo
// add ipv6

type Registry struct {
	name     string
	mu       sync.RWMutex
	logger   *slog.Logger
	resolver *zeroconf.Resolver
}

func NewRegistry(namespace *Namespace) (*Registry, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, err
	}

	logger := namespace.Logger().With("registry", namespace.Name())

	reg := &Registry{
		name:     namespace.Name(),
		resolver: resolver,
		logger:   logger,
	}

	return reg, nil
}

func (r *Registry) Lookup(ctx context.Context, name string) (string, error) {

	// Use a buffer of 1 to prevent goroutine leaks
	entries := make(chan *zeroconf.ServiceEntry, 1)
	if err := r.resolver.Lookup(ctx, name, "_"+r.name+globals.ZERO_CONF_NODE_TYPE, globals.ZERO_CONF_DOMAIN, entries); err != nil {
		return "", err
	}

	select {
	case entry := <-entries:
		if entry == nil {
			return "", errors.New("service entry is nil (channel closed with no results)")
		}
		if len(entry.AddrIPv4) > 0 {
			return net.JoinHostPort(entry.AddrIPv4[0].String(), strconv.Itoa(entry.Port)), nil
		}

		if len(entry.AddrIPv6) > 0 {
			return net.JoinHostPort(entry.AddrIPv6[0].String(), strconv.Itoa(entry.Port)), nil
		}

		return "", errors.New("no IP address found for service")
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
