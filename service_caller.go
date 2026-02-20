package spine

import (
	"context"
	"fmt"

	"github.com/poisnoir/mad-go"
)

type ServiceCaller[K any, V any] struct {
	namespace    *Namespace
	serviceName  string
	keyEncoder   *mad.Mad[K]
	valueEncoder *mad.Mad[V]
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewServiceCaller[K any, V any](namespace *Namespace, serviceName string, ctx context.Context) (*ServiceCaller[K, V], error) {

	keyEnc, err := mad.NewMad[K]()
	if err != nil {
		return nil, err
	}

	valueEnc, err := mad.NewMad[V]()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(namespace.ctx)
	return &ServiceCaller[K, V]{
		namespace:    namespace,
		keyEncoder:   keyEnc,
		valueEncoder: valueEnc,
		serviceName:  serviceName,
		ctx:          ctx,
		cancel:       cancel,
	}, nil
}

func (sc *ServiceCaller[K, V]) Call(key K, ctx context.Context) (V, error) {
	var response V

	ipAddress, err := sc.namespace.GetService(sc.serviceName, ctx)
	if err != nil {
		return response, err
	}

	bufPtr := sc.namespace.bufferPool.Get().(*[]byte)
	payload := (*bufPtr)[:sc.keyEncoder.GetRequiredSize(&key)]
	defer sc.namespace.bufferPool.Put(bufPtr)

	err = sc.keyEncoder.Encode(&key, payload)
	if err != nil {
		return response, fmt.Errorf("failed to encode key: %w", err)
	}

	err = request(ctx, ipAddress, &payload, sc.namespace.encryption, []byte(sc.namespace.secretKey))
	if err != nil {
		return response, fmt.Errorf("failed to send request: %w", err)
	}

	err = sc.valueEncoder.Decode(payload, &response)
	if err != nil {
		return response, fmt.Errorf("failed to decode response: %w", err)
	}

	return response, nil
}

func (sc *ServiceCaller[K, V]) Close() {
	sc.cancel()
}
