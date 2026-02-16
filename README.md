# spine

`spine` is a high-performance, decentralized communication library for robotics and IoT systems. 
It provides ROS-like communication patterns—Services (RPC) and Pub/Sub—built natively in Go with zero external brokers, minimal overhead, and absolute developer freedom.

---
# Architecture
spine treats every component as a standalone Entity. 
Whether you are running a single monolithic binary or a distributed swarm of microservices, discovery and connectivity are handled transparently at the library level.
- `Service/ServiceCaller`: Synchronous Request/Response.
- `Publishers/Subscribers`: Asynchronous broadcast.
- `Streamer`: Asynchronous streaming.
- Namespaces: Virtual silos that group related services and prevent cross-talk.

All components are automatically discovered on the local network using `zeroconf`.

---
# Getting Started

### 1. Join a Namespace
All communication happens within a secured namespace.

```go
    logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
    // Join a namespace with a secret key and encryption disabled
    // If encryption is false, hmac is used for authentication
    ns, err := spine.JointNamespace("mecca500", "secret_meow", logger, false)
```

### 2. Create a Service
Turn any Go function into a network-discoverable service.
```go
    // Define a handler: func(InputType) (OutputType, error)
    handler := func(input string) (int, error) {
        return len(input), nil
    }

    // Register the service
    _, err = spine.NewService(ns, "string_length", handler)
```

### 3. Call a Service
Call services from any machine on the network using the same generic types.
```go
    ctx, _ := context.WithTimeout(context.Background(), time.Second)

    // will error if types are mismatched
    // Blocks until result is received or context is canceled
    c, err := NewServiceCaller[string, int](ns, "string_length", ctx)
	
		// will error if it can't get result before context cancels
		// Blocks until result is received or context is canceled
		result, err := c.Call("hello world", ctx)
```

---
## Pub/Sub (Beta)
Publishers and Subscribers allow for asynchronous data flow. Connections are established automatically once a publisher is discovered on the network.
```go
    // Create a Publisher
    pub, _ := spine.NewPublisher[SensorData]("lidar_scan")
    pub.Publish(currentData)

    // Create a Subscriber
    sub, _ := spine.NewSubscriber[SensorData]("lidar_scan", func(data SensorData) {
    fmt.Printf("Received data: %v\n", data)
})
```

# Dependencies
- zeroconf https://github.com/grandcat/zeroconf: Service discovery
- mad-go https://github.com/poisnour/mad-go: Serialization
- kcp-go https://github.com/xtaci/kcp-go: Network Protocol

# Contribution
Feel free to contribute or suggest features. Contact: @rima1881
