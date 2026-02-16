package spine

import (
	"context"
	"fmt"
	"slices"

	"github.com/poisnoir/mad-go"
)

type ServiceCaller[K any, V any] struct {
	namespace    *Namespace
	serviceName  string
	keyEncoder   *mad.Mad[K]
	valueEncoder *mad.Mad[V]
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

	ipAddress, err := namespace.GetService(serviceName, ctx)
	if err != nil {
		return nil, err
	}

	bufPtr := namespace.bufferPool.Get().(*[]byte)
	tmpBuf := (*bufPtr)[:0]
	defer func() {
		*bufPtr = tmpBuf
		namespace.bufferPool.Put(bufPtr)
	}()

	payload := []byte(keyEnc.Code())
	if namespace.enyption == nil {
		signature := generateHmac([]byte(namespace.secretKey), payload)
		tmpBuf = append(tmpBuf, signature...)
	}
	tmpBuf = append(tmpBuf, payload...)

	err = request(ctx, ipAddress, &tmpBuf, namespace.enyption)
	if err != nil {
		return nil, err
	}

	if namespace.enyption == nil {
		if len(tmpBuf) < 32 {
			return nil, fmt.Errorf("response too short")
		}
		sig, data := tmpBuf[:32], tmpBuf[32:]
		if !verifyHmac([]byte(namespace.secretKey), data, sig) {
			return nil, fmt.Errorf("corrupted response: HMAC mismatch")
		}
		tmpBuf = data
	}

	if slices.Equal(tmpBuf, []byte(valueEnc.Code())) {
		return nil, fmt.Errorf("service layout does not match service listener layout")
	}

	return &ServiceCaller[K, V]{
		namespace:    namespace,
		keyEncoder:   keyEnc,
		valueEncoder: valueEnc,
		serviceName:  serviceName,
	}, nil
}

func (sc *ServiceCaller[K, V]) Call(key K, ctx context.Context) (V, error) {
	var response V

	ipAddress, err := sc.namespace.GetService(sc.serviceName, ctx)
	if err != nil {
		return response, err
	}

	ptr := sc.namespace.bufferPool.Get().(*[]byte)
	buff := *ptr
	err = sc.keyEncoder.Encode(&key, buff)
	if err != nil {
		return response, fmt.Errorf("payload is too large, max service payload size is 4kb")
	}

	// Todo: Add Backoff mechanism
	err = request(ctx, ipAddress, ptr, sc.namespace.enyption)
	if err != nil {
		return response, err
	}
	// err = sc.valueEncoder.Decode(buff[:bytesRead], &response)
	if err != nil {
		return response, fmt.Errorf("failed to decode response: %w", err)
	}

	return response, nil
}
