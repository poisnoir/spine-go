package spine

import (
	"context"
	"fmt"

	"github.com/xtaci/kcp-go/v5"
)

func request(ctx context.Context, ipAddress string, buf *[]byte, encrypt kcp.BlockCrypt, key []byte) error {

	if encrypt == nil {
		signature := generateHmac(key, *buf)
		(*buf) = append(*buf, signature...)
	}

	sess, err := kcp.DialWithOptions(ipAddress, encrypt, 10, 3)
	if err != nil {
		return err
	}

	stop := context.AfterFunc(ctx, func() {
		sess.Close()
	})
	defer stop()

	_, err = sess.Write(*buf)
	if err != nil {
		return err
	}

	*buf = (*buf)[:0]
	n, err := sess.Read(*buf)
	if err != nil {
		return err
	}

	if encrypt == nil {
		if len(*buf) < 32 {
			return fmt.Errorf("response too short")
		}
		data, sig := (*buf)[:len(*buf)-32], (*buf)[len(*buf)-32:]
		if !verifyHmac([]byte(key), data, sig) {
			return fmt.Errorf("corrupted response: HMAC mismatch")
		}
		(*buf) = data
	}

	*buf = (*buf)[:n]

	return nil
}
