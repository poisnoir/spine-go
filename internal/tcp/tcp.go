package tcp

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/Pois-Noir/Botzilla/internal/globals"
)

func Request(ctx context.Context, destinationAddress string, payload []byte) ([]byte, error) {

	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", destinationAddress)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		conn.SetDeadline(deadline)
	} else {
		return nil, context.DeadlineExceeded
	}

	requestHeader := NewHeader(globals.OK_STATUS, payload)
	encodedHeader := requestHeader.Encode()

	message := append(encodedHeader, payload...)
	_, err = conn.Write(message)
	if err != nil {
		return nil, fmt.Errorf("failed to write message: %w", err)
	}

	var responseHeaderBuffer [globals.HEADER_LENGTH]byte
	_, err = io.ReadFull(conn, responseHeaderBuffer[:])
	if err != nil {
		return nil, fmt.Errorf("failed to read response header: %w", err)
	}

	responseHeader, err := Decode(responseHeaderBuffer[:])
	if err != nil {
		return nil, err
	}

	rawResponse := make([]byte, responseHeader.PayloadLength)
	_, err = io.ReadFull(conn, rawResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return rawResponse, nil
}
