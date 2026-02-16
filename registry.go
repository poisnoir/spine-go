package spine

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/poisnoir/spine-go/internal/globals"
)

// Todo
// add ipv6

type component struct {
	addr   string
	expiry time.Time
}

type Registry struct {
	name       string
	mu         sync.RWMutex
	components map[string]component
	logger     *slog.Logger
	resolver   *zeroconf.Resolver
}

func NewRegistry(namespace *Namespace) (*Registry, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, err
	}

	logger := namespace.Logger().With("registry", namespace.Name())

	reg := &Registry{
		components: make(map[string]component),
		name:       namespace.Name(),
		resolver:   resolver,
		logger:     logger,
	}

	go reg.start(namespace.Context())
	go reg.startJanitor(namespace.Context(), 30*time.Second)

	return reg, nil
}

func (r *Registry) start(ctx context.Context) {
	entries := make(chan *zeroconf.ServiceEntry)
	go r.resolver.Browse(ctx, r.name+globals.ZERO_CONF_TYPE, globals.ZERO_CONF_DOMAIN, entries)
	for {
		select {
		case <-ctx.Done():
			r.logger.Info("Discovery stopped")
			return
		case entry := <-entries:
			r.mu.Lock()
			r.components[entry.Instance] = component{
				addr:   entry.AddrIPv4[0].String() + ":" + strconv.Itoa(entry.Port),
				expiry: time.Now().Add(time.Duration(entry.TTL) * time.Second),
			}
			r.mu.Unlock()
		}
	}
}

func (r *Registry) Lookup(ctx context.Context, name string) (string, error) {
	r.mu.RLock()
	comp, exists := r.components[name]
	r.mu.RUnlock()

	if exists && time.Now().Before(comp.expiry) {
		return comp.addr, nil
	}

	// Use a buffer of 1 to prevent goroutine leaks
	entries := make(chan *zeroconf.ServiceEntry, 1)
	if err := r.resolver.Lookup(ctx, name, r.name+globals.ZERO_CONF_TYPE, globals.ZERO_CONF_DOMAIN, entries); err != nil {
		return "", err
	}

	select {
	case entry := <-entries:
		if len(entry.AddrIPv4) == 0 {
			return "", errors.New("no IPv4 address found for service")
		}
		addr := net.JoinHostPort(entry.AddrIPv4[0].String(), strconv.Itoa(entry.Port))

		return addr, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func (r *Registry) startJanitor(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.mu.Lock()
			now := time.Now()
			for name, comp := range r.components {
				if now.After(comp.expiry) {
					r.logger.Debug("Cleaning up expired component", "name", name)
					delete(r.components, name)
				}
			}
			r.mu.Unlock()
		}
	}
}

func (r *Registry) RemoveOnFailure(failedComp string) {
	r.mu.Lock()
	delete(r.components, failedComp)
	r.mu.Unlock()
}
