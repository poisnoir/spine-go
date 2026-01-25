package tcp

import (
	"encoding/binary"
	"errors"

	"github.com/Pois-Noir/Botzilla/internal/globals"
)

type Header struct {
	Status        uint8
	PayloadLength uint32
}

func NewHeader(status uint8, payload []byte) *Header {
	return &Header{
		Status:        status,
		PayloadLength: uint32(len(payload)),
	}
}

func (h *Header) Encode() []byte {

	buf := make([]byte, globals.HEADER_LENGTH)

	buf[globals.STATUS_CODE_INDEX] = h.Status
	binary.BigEndian.PutUint32(buf[globals.MESSAGE_LENGTH_INDEX:], h.PayloadLength)

	return buf
}

func Decode(buffer []byte) (*Header, error) {

	if len(buffer) != globals.HEADER_LENGTH {
		return nil, errors.New("buffer size is not correct")
	}

	header := &Header{}
	header.Status = buffer[globals.STATUS_CODE_INDEX]
	header.PayloadLength = binary.BigEndian.Uint32(buffer[globals.MESSAGE_LENGTH_INDEX:globals.HEADER_LENGTH])

	return header, nil

}
