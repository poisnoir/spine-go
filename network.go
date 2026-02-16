package spine

import (
	"context"

	"github.com/xtaci/kcp-go/v5"
)

func request(ctx context.Context, ipAddress string, buf *[]byte, encrypt kcp.BlockCrypt) error {

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

	n, err := sess.Read(*buf)
	if err != nil {
		return err
	}
	*buf = (*buf)[:n]

	return nil
}
