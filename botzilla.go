package botzilla

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/Pois-Noir/Botzilla/internal/globals"
	"github.com/Pois-Noir/Botzilla/internal/registry"
)

type botzilla struct {
	useEncryption bool   // It is gonna slow down all comm types.
	secretKey     string // key used for generating hmac. Do not share it in the network!!!
	reg           *registry.Registry
	cancel        context.CancelFunc
	logger        *slog.Logger
}

var (
	b  *botzilla
	mu sync.RWMutex // Protects the global instance 'b'
)

func Start(useEncryption bool, secretKey string, logger *slog.Logger) error {
	mu.Lock()
	defer mu.Unlock()

	if b != nil {
		return fmt.Errorf("botzilla is already running")
	}

	ctx, cancel := context.WithCancel(context.Background())

	reg, err := registry.NewRegistry(ctx)
	if err != nil {
		cancel()
		return err
	}

	b = &botzilla{
		useEncryption: useEncryption,
		secretKey:     secretKey,
		reg:           reg,
		logger:        logger,
		cancel:        cancel,
	}

	go reg.StartBrowser()
	return nil
}

func Stop() {
	mu.Lock()
	defer mu.Unlock()

	if b != nil {
		b.cancel()
		b = nil // Clear it so it can be started again if needed
	}
}

func GetService(cxt context.Context, name string) (string, error) {
	return b.reg.Lookup(cxt, name+globals.ZERO_CONF_SERVICE_PREFIX)
}

func GetPublisher(ctx context.Context, name string) (string, error) {
	return b.reg.Lookup(ctx, name+globals.ZERO_CONF_PUBLISHER_PREFIX)
}

func GetLogger() *slog.Logger {
	mu.RLock()
	defer mu.RUnlock()
	return b.logger
}
