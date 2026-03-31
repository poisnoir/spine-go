package spine

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"slices"

	"github.com/grandcat/zeroconf"
	"github.com/poisnoir/mad-go"
	"github.com/poisnoir/spine-go/internal/globals"
	"github.com/xtaci/kcp-go/v5"
)

// bunch of same operations in service and threaded service

func generateService[K any, V any](namespace *Namespace, name string) (*mad.Mad[K], *mad.Mad[V], *kcp.Listener, *zeroconf.Server, error) {
	logger := namespace.logger.With(
		namespace.Name(),
		"service",
		name,
		"new service",
	)

	keyEnc, err := mad.NewMad[K]()
	if err != nil {
		logger.Error("unable to create key encoder", "error", err)
		return nil, nil, nil, nil, err
	}

	valueEnc, err := mad.NewMad[V]()
	if err != nil {
		logger.Error("unable to create value encoder", "error", err)
		return nil, nil, nil, nil, err
	}

	listener, err := kcp.ListenWithOptions(":0", namespace.encryption, 10, 3)
	if err != nil {
		logger.Error("unable to create listener", "error", err)
		return nil, nil, nil, nil, err
	}

	server, err := zeroconf.Register(
		globals.ZERO_CONF_SERVICE_PREFIX+name,
		namespace.Name()+globals.ZERO_CONF_TYPE,
		globals.ZERO_CONF_DOMAIN,
		listener.Addr().(*net.UDPAddr).Port,
		[]string{"id=spine_service_" + name},
		nil,
	)

	if err != nil {
		logger.Error("unable to register service to zeroconf", "error", err)
		return nil, nil, nil, nil, err
	}

	return keyEnc, valueEnc, listener, server, nil

}

func establishConnection(conn io.ReadWriteCloser, keyCode []byte, valueCode []byte, buf []byte, logger *slog.Logger) error {
	n, err := conn.Read(buf)
	if err != nil {
		return err
	}

	if !slices.Equal(keyCode, buf[:n]) {
		logger.Error("failed to establish connection")
		return fmt.Errorf("invalid key code")
	}

	_, err = conn.Write([]byte{globals.OK_STATUS_CODE})
	if err != nil {
		logger.Error("failed to establish connection")
		return err
	}

	n, err = conn.Read(buf)
	if err != nil {
		return err
	}

	if !slices.Equal(valueCode, buf[:n]) {
		logger.Error("failed to establish connection")
		return fmt.Errorf("invalid value code")
	}
	_, err = conn.Write([]byte{globals.OK_STATUS_CODE})

	return err
}

func handleCallerRequest[K any, V any](conn io.ReadWriteCloser, keySerializer *mad.Mad[K], valueSerializer *mad.Mad[V], errorSerializer *mad.Mad[string], buf []byte, processRequest func(K) serviceOutput[V], logger *slog.Logger) {

	defer conn.Close()
	err := establishConnection(conn, []byte(keySerializer.Code()), []byte(valueSerializer.Code()), buf, logger)
	if err != nil {
		return
	}

	for {
		n, err := conn.Read(buf)
		if err != nil {
			logger.Error("unable to read from connection", "error", err)
			return
		}

		if buf[0] == globals.PING_CODE {
			conn.Write([]byte{globals.PONG_CODE})
			continue
		}

		// might change in future if we add more operation codes
		if buf[0] != globals.SERVICE_REQUEST {
			logger.Error("received invalid operation code", "error", err)
			conn.Write([]byte{globals.ERROR_INVALID_OPERATION_CODE})
			continue
		}

		var key K
		err = keySerializer.Decode(buf[1:n], &key)
		if err != nil {
			logger.Error("unable to decode key", "error", err)
			continue
		}

		res := processRequest(key)
		if res.err != nil {
			logger.Error("handler failed", "error", err)
			errMsg := res.err.Error()
			buf[0] = globals.ERROR_SERVICE_ERROR_CODE
			errorSerializer.Encode(&errMsg, buf[1:])
			_, err = conn.Write(buf)
			if err != nil {
				logger.Error("failed to write from connection", "error", err)
				return
			}
		}

		valueSerializer.Encode(&res.data, buf)
		_, err = conn.Write(buf[:valueSerializer.GetRequiredSize(&res.data)])
		if err != nil {
			logger.Error("failed to write from connection", "error", err)
			return
		}

	}

}
