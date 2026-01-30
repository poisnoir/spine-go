package udp

import (
	"context"
	"log/slog"
	"net"
)

func Request(ctx context.Context, udpAddr *net.UDPAddr, payload []byte, rawResponse []byte) error {

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write(payload)
	if err != nil {
		return err
	}

	_, err = conn.Read(rawResponse)
	return err
}

type Listener[K any, V any] struct {
	tcpListener net.Listener
	handler     func(K) (V, error)
	sem         chan int
	logger      *slog.Logger
	context     context.Context
}
