package spine

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/poisnoir/spine-go/internal/globals"
	"github.com/xtaci/kcp-go/v5"
)

// The buffer is coming from pool and it is big enoght
// the response is gonna go to given buffer if exist
func write(sess io.ReadWriteCloser, buf []byte, requestSize int, hasResponse bool) (int, error) {

	_, err := sess.Write(buf[:requestSize])
	if !hasResponse {
		return 0, err
	}

	n, err := sess.Read(buf)
	return n, err
}

func runListener(listener *kcp.Listener, logger *slog.Logger, handler func(io.ReadWriteCloser)) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("unable to accept connection", "error", err)
			continue
		}
		go handler(conn)
	}
}

func ping(conn io.ReadWriteCloser) error {
	buf := []byte{globals.PING_CODE}
	_, err := write(conn, buf, 1, true)
	if err != nil {
		return err
	}

	if buf[0] != globals.PONG_CODE {
		return fmt.Errorf(globals.ERROR_PING)
	}

	return nil
}

func handlerReceiver(conn io.ReadWriteCloser, handler func([]byte)) {

	for {

	}

}
