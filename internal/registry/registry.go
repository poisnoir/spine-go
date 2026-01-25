package registry

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/Pois-Noir/Botzilla/internal/globals"
	"github.com/grandcat/zeroconf"
)

// Todo
// remove old dead services

type Registry struct {
	mu         sync.RWMutex
	components map[string]string
	ctx        context.Context
	resolver   *zeroconf.Resolver
}

func NewRegistry(ctx context.Context) (*Registry, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, err
	}

	reg := &Registry{
		components: make(map[string]string),
		resolver:   resolver,
		ctx:        ctx,
	}

	return reg, nil
}

func (r *Registry) StartBrowser() {
	entries := make(chan *zeroconf.ServiceEntry)
	go r.resolver.Browse(r.ctx, globals.ZERO_CONF_SERVICE, globals.ZERO_CONF_DOMAIN, entries)

	for {

		select {
		case <-r.ctx.Done():
			// Todo
			// change it with real log
			fmt.Println("Discovery stopped!!!")
			return
		case entry := <-entries:
			r.mu.Lock()
			r.components[entry.Instance] = entry.AddrIPv4[0].String() + ":" + strconv.Itoa(entry.Port)
			r.mu.Unlock()

		}
	}

}

func (r *Registry) Lookup(ctx context.Context, name string) (string, error) {

	r.mu.RLock()
	addr := r.components[name]
	r.mu.RUnlock()

	if addr != "" {
		return addr, nil
	}

	// do a lookup if component is not registered yet
	// this is a special case when the component has just created and browser hasn't registered component yet
	entries := make(chan *zeroconf.ServiceEntry)
	err := r.resolver.Lookup(ctx, name, globals.ZERO_CONF_SERVICE, globals.ZERO_CONF_DOMAIN, entries)
	if err != nil {
		return "", err
	}

	select {
	case entry := <-entries:
		// Todo
		// need to add ipv6 support
		return entry.AddrIPv4[0].String() + ":" + strconv.Itoa(entry.Port), nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}
