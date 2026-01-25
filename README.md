# Botzilla

`Botzilla` is a lightweight communication library for robotics and IoT systems.
It enables service calls and pub/sub messaging with zero configuration, minimal dependencies, and strong typing using Go generics.

Designed for developers who want:
- ROS-like communication patterns
- Without heavy frameworks
- Without brokers
- Without complex setup

**Important: This library is still on beta. Some features are still not production ready grade**

---
# Architecture
Botzilla is built around three core concepts:
- `Services` – Request/response communication
- `Publishers` – Broadcast messages
- `Subscribers` – Async message handlers

All components are automatically discovered on the local network using `zeroconf`.

---
## Start/Stop
Botzilla uses a couple of background services and settings. before using the library Start function must be called
```go
    func Start(useEncryption bool, secretKey string, logger *slog.Logger) error
```

Stoping all background services of library
```go
    func Stop()
```

## Services
Services represent API endpoints using a request/response model.
Each service runs on a local TCP server and is discoverable by name.

### create a Service

```go
	func NewService[K any, V any](name string, maxHandler uint32, handler func(K) (V, error)) (*Service[V, K], error) 
```
K type is the input value of service and V is output value. 

#### Parameters
`name` : Unique service identifier on the network.

`maxhandler` : Maximum number of concurrent requests. When the limit is reached, new requests wait.
**Important**: In order to ensure memory safety it is recommended to use `1` for maxHandler or make sure your handler does
not introduce race conditions.

`handler`: Function executed on each request.

---

### Call a service

```go
	func Call[K any, V any](serviceName string, payload K, timeout time.Duration) (V, error)
```
sends a request to service.

---

## Publishers
Publishers broadcast data without expecting a response.
```go
	func NewPublisher[K any](name string) (*Publisher[K], error)
```

```go
	func (p *Publisher[K]) Publish(data K) error 
```
sends data to all subscribers.

## Subscribers
Subscribers listen to publishers and run a handler when data arrives.
`Note`: The publisher does not need to exist when the subscriber starts. Once the publisher appears, the connection is automatic.
```go
	func NewSubscriber[K any](publisherName string, handler func(K)) (*Subscriber[K], error)
```
```go
	func (subscriber *Subscriber[K]) Stop() error
```
## Discovery Helpers

```go
	func GetService(name string) (string, error)
```

```go
	func GetAllServices() (map[string]string, error)
```

```go
	func GetPublisher(name string) (string, error)
```

```go
	func GetAllPublishers() (map[string]string, error)
```


# Security
To secure incoming messages, **HMAC** (Hash-based Message Authentication Code) with a shared secret key is used. This ensures the integrity and authenticity of the messages.

**Note:** The messages themselves are not encrypted, meaning that while HMAC ensures authenticity, the content of the messages could still be exposed if intercepted. Encryption for message content is planned for future releases.


# Dependencies
- zeroconf https://github.com/grandcat/zeroconf: Service discovery
- MessagePack: Binary serialization

# Contribution
Feel free to contribute or suggest features. Contact: @rima1881