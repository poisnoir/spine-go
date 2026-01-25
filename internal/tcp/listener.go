package tcp

import (
	"bufio"
	"errors"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/Pois-Noir/Botzilla/internal/globals"
	"github.com/Pois-Noir/Botzilla/internal/utils"

	"github.com/vmihailenco/msgpack/v5"
)

type Listener[K any, V any] struct {
	tcpListener net.Listener
	handler     func(K) (V, error)
	isAlive     *utils.SafeVar[bool]
	sem         chan int
	logger      *slog.Logger
}

func NewListener[K any, V any](handler func(K) (V, error), maxHandlers uint32, logger *slog.Logger) (*Listener[K, V], error) {

	tcpListener, err := net.Listen("tcp", ":0") // open on the random free port
	if err != nil {
		return nil, err
	}

	return &Listener[K, V]{
		tcpListener: tcpListener,
		handler:     handler,
		isAlive:     utils.NewSafeVar(true),
		sem:         make(chan int, maxHandlers),
		logger:      logger,
	}, nil
}

func (l *Listener[K, V]) Start() {

	l.logger.Info("listener starting", "port", l.Port(), "max_concurrency", cap(l.sem))

	for i := 0; i < cap(l.sem); i++ {
		l.sem <- i
	}

	for {
		conn, err := l.tcpListener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				l.logger.Info("listener closed")
				break
			}
			l.logger.Error("accept error", "error", err)
			continue
		}

		i := <-l.sem
		go l.handleRequest(conn, i)
	}

	// wait for all handlers to finish
	for i := 0; i < cap(l.sem); i++ {
		_ = <-l.sem
	}

	l.isAlive.Set(false)
}

func (l *Listener[K, V]) Kill() error {
	return l.tcpListener.Close()
}

func (l *Listener[K, V]) Port() int {
	addr := l.tcpListener.Addr().(*net.TCPAddr)
	return addr.Port
}

func (l *Listener[K, V]) handleRequest(conn net.Conn, num int) {

	remoteAddr := conn.RemoteAddr().String()
	// Create a logger for this specific connection
	log := l.logger.With("remote_addr", remoteAddr, "worker_id", num)

	defer conn.Close()

	// letting others know handler has finished
	defer func() {
		l.sem <- num
	}()

	// Todo
	// somehow change this bitch
	err := conn.SetDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		log.Warn("failed to set connection deadline", "error", err)
		return
	}

	// creating a buffered Reader
	bReader := bufio.NewReader(conn)

	var requestHeaderBuffer [globals.HEADER_LENGTH]byte
	_, err = io.ReadFull(bReader, requestHeaderBuffer[:])

	if err != nil {
		log.Warn("failed to read header", "error", err)
		return
	}

	requestHeader, err := Decode(requestHeaderBuffer[:])
	if err != nil {
		log.Warn("failed to decode header", "error", err)
		return
	}

	// get the message length
	requestSize := requestHeader.PayloadLength
	requestPayloadBuffer := make([]byte, requestSize)

	// We use io.ReadFull to guarantee that we read exactly `requestSize` bytes.
	_, err = io.ReadFull(bReader, requestPayloadBuffer)

	if err != nil {
		log.Warn("failed to read payload", "size", requestHeader.PayloadLength, "error", err)
		return
	}

	var requestPayload K
	err = msgpack.Unmarshal(requestPayloadBuffer[:], &requestPayload)

	if err != nil {
		log.Warn("msgpack unmarshal failed", "error", err)
		return
	}

	// run user callback
	start := time.Now()
	responsePayload, err := l.handler(requestPayload)
	duration := time.Since(start)

	// send error response code
	if err != nil {
		log.Error("handler logic error", "error", err, "duration", duration)
		return
	}

	responsePayloadBuffer, err := msgpack.Marshal(responsePayload)

	// logger
	if err != nil {
		log.Error("failed to marshal response", "error", err)
		return
	}

	responseHeader := NewHeader(
		globals.OK_STATUS,
		responsePayloadBuffer,
	)
	responseHeaderBuffer := responseHeader.Encode()

	responseBuffer := append(responseHeaderBuffer, responsePayloadBuffer...)
	_, err = conn.Write(responseBuffer)

	if err != nil {
		log.Warn("failed to write response to client", "error", err)
		return
	}

	log.Debug("request processed successfully", "duration", duration)

}

func (l *Listener[K, V]) IsAlive() bool {
	return l.isAlive.Get()
}
