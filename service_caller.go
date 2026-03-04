package spine

import (
	"context"
	"fmt"

	"github.com/poisnoir/mad-go"
	"github.com/poisnoir/spine-go/internal/globals"
	"github.com/xtaci/kcp-go/v5"
)

type kcpPool struct {
}

type ServiceCaller[K any, V any] struct {
	namespace    *Namespace
	serviceName  string
	keyEncoder   *mad.Mad[K]
	valueEncoder *mad.Mad[V]
	conn         *kcp.UDPSession
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewServiceCaller[K any, V any](namespace *Namespace, serviceName string) (*ServiceCaller[K, V], error) {

	keyEnc, err := mad.NewMad[K]()
	if err != nil {
		return nil, err
	}

	valueEnc, err := mad.NewMad[V]()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(namespace.ctx)

	sc := &ServiceCaller[K, V]{
		namespace:    namespace,
		keyEncoder:   keyEnc,
		valueEncoder: valueEnc,
		serviceName:  serviceName,
		ctx:          ctx,
		cancel:       cancel,
	}

	return sc, nil
}

/*
The connection routine attempts to establish a KCP connection to the service.
If the connection drops, this function is called to reconnect. Given that we expect a maximum of 50 clients per service,
maintaining persistent connections should not lead to port exhaustion. However, if port exhaustion becomes an issue in the future,
we will implement a mechanism to disconnect clients with low load.
*/
func (sc *ServiceCaller[K, V]) connect(ctx context.Context) error {

	logger := sc.namespace.logger.With(
		sc.namespace.Name(),
		"service_caller",
		sc.serviceName,
		"connect",
	)

	// finding the service
	address, err := sc.namespace.GetService(sc.serviceName, ctx)
	if err != nil {
		logger.Error("unable to find the service", "error", err)
		return err // the only way to fail here is to run out of context
	}

	// establishing connection
	sess, err := kcp.DialWithOptions(address, sc.namespace.encryption, 10, 3)
	if err != nil {
		logger.Error("failed to dial service", "error", err)
		return err
	}

	// getting buffer for comm
	bufPtr := sc.namespace.bufferPool.Get().(*[]byte)
	defer sc.namespace.bufferPool.Put(bufPtr)
	buf := *bufPtr

	// validating input/output service types
	keyCode := sc.keyEncoder.Code()
	n := copy(buf, keyCode)

	n, err = request(sess, buf[:n], true)
	if err != nil {
		logger.Error("failed to validate service input type", "error", err)
		return err
	} else if n != 1 {
		err = fmt.Errorf("response is corrupted")
		logger.Error("failed to validate service input type", "error", err)
		return err
	} else if buf[0] != globals.OK_STATUS {
		err = fmt.Errorf("service data type is different")
		logger.Error("failed to validate service input type", "error", err)
		return err
	}

	valueCode := sc.valueEncoder.Code()
	n = copy(buf, valueCode)

	n, err = request(sess, buf[:n], true)
	if err != nil {
		logger.Error("failed to validate service output type", "error", err)
	} else if n != 1 {
		err = fmt.Errorf("response is corrupted")
		logger.Error("failed to validate service output type", "error", err)
		return err
	} else if buf[0] != globals.OK_STATUS {
		err = fmt.Errorf("service data type is different")
		logger.Error("failed to validate service output type", "error", err)
		return err
	}
	sc.conn = sess

	return nil

}

// sends data: key to the service and returns V from service
// context is used for establishing connection
func (sc *ServiceCaller[K, V]) Call(key K, ctx context.Context) (V, error) {
	var v V

	var err error = nil
	if err != nil {
		return v, err
	}

	requestSize := sc.keyEncoder.GetRequiredSize(&key)
	if requestSize > globals.MAX_PACKET_SIZE {
		return v, fmt.Errorf("failed to encode key. key is too big. max key size is 4kb")
	}

	bufPtr := sc.namespace.bufferPool.Get().(*[]byte)
	defer sc.namespace.bufferPool.Put(bufPtr)
	buf := *bufPtr
	sc.keyEncoder.Encode(&key, buf)

	_, err = request(sc.conn, buf[:requestSize], true)
	if err != nil {
		return v, err
	}

	_ = sc.valueEncoder.Decode(buf, &v)
	return v, nil
}

func (sc *ServiceCaller[K, V]) Close() {
	sc.cancel()
}
