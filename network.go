package spine

import (
	"github.com/xtaci/kcp-go/v5"
)

// The buffer is coming from pool and it is big enoght
// the response is gonna go to given buffer if exist
func write(sess *kcp.UDPSession, buf []byte, requestSize int, hasResponse bool) (int, error) {

	_, err := sess.Write(buf[:requestSize])
	if !hasResponse {
		return 0, err
	}

	n, err := sess.Read(buf)
	return n, err
}
