package spine

import (
	"github.com/xtaci/kcp-go/v5"
)

// The buffer is coming from pool and it is big enoght
// the response is gonna go to given buffer if exist
func request(sess *kcp.UDPSession, buf []byte, hasResponse bool) (int, error) {

	_, err := sess.Write(buf)
	if !hasResponse {
		return 0, err
	}

	buf = buf[:0]
	n, err := sess.Read(buf)
	return n, err
}
