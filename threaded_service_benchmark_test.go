package spine

import (
	"context"
	"io"
	"log/slog"
	"testing"
)

func BenchmarkThreadedServiceCall(b *testing.B) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ns, err := JointNamespace("bench_threaded", "secret", logger, false)
	if err != nil {
		b.Fatal(err)
	}

	handler := func(input uint32) (uint32, error) {
		return input * 2, nil
	}

	_, err = NewThreadedService(ns, "math_threaded", handler)
	if err != nil {
		b.Fatal(err)
	}

	caller, err := NewServiceCaller[uint32, uint32](ns, "math_threaded")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := caller.Call(uint32(i), context.Background())
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkThreadedServiceCallEncrypted(b *testing.B) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ns, err := JointNamespace("bench_threaded_encrypted", "secret", logger, true)
	if err != nil {
		b.Fatal(err)
	}

	handler := func(input uint32) (uint32, error) {
		return input * 2, nil
	}

	_, err = NewThreadedService(ns, "math_threaded_enc", handler)
	if err != nil {
		b.Fatal(err)
	}

	caller, err := NewServiceCaller[uint32, uint32](ns, "math_threaded_enc")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := caller.Call(uint32(i), context.Background())
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkThreadedServiceCallParallel(b *testing.B) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ns, err := JointNamespace("bench_threaded_parallel", "secret", logger, false)
	if err != nil {
		b.Fatal(err)
	}

	handler := func(input uint32) (uint32, error) {
		return input * 2, nil
	}

	_, err = NewThreadedService(ns, "math_threaded_parallel", handler)
	if err != nil {
		b.Fatal(err)
	}

	caller, err := NewServiceCaller[uint32, uint32](ns, "math_threaded_parallel")
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := uint32(0)
		for pb.Next() {
			_, err := caller.Call(i, context.Background())
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}
