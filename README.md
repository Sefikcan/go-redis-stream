# go-redis-stream

A sample Go project demonstrating how to use **Redis Streams** for building distributed producer-consumer systems. This project provides a simple order processing pipeline using Redis Streams, with separate producer and consumer applications.

## What is Redis Streams?

[Redis Streams](https://redis.io/docs/data-types/streams/) is a powerful data structure introduced in Redis 5.0, designed for managing real-time data flows. It enables message brokering, event sourcing, and log-like data storage, supporting features like message persistence, consumer groups, and at-least-once delivery semantics.

**Key features of Redis Streams:**
- Append-only log structure for messages
- Consumer groups for scalable message processing
- Message acknowledgment and delivery tracking
- Persistence and durability

## Project Structure

- `cmd/producer/`: Producer application that generates and sends order messages to a Redis stream.
- `cmd/consumer/`: Consumer application that reads and processes order messages from the Redis stream.
- `internal/stream/`: Core logic for interacting with Redis Streams (produce/consume).
- `internal/order/`: Order model and handler logic.
- `config/`: Configuration loader for environment variables.

## Data Model

The system processes simple order messages with the following structure:

```go
// internal/order/model.go
 type Order struct {
   ID     string `json:"id"`
   Item   string `json:"item"`
   Amount int    `json:"amount"`
 }
```

## Getting Started

### Prerequisites
- [Go 1.23+](https://golang.org/dl/)
- [Docker](https://www.docker.com/) (for running Redis)

### Running Redis with Docker

A `docker-compose.yml` is provided:

```yaml
version: "3"
services:
  redis:
    image: redis:6.2
    ports:
      - "6379:6379"
```

Start Redis:

```sh
docker-compose up -d
```

### Configuration

The applications use the following environment variable:
- `REDIS_ADDR` (default: `localhost:6379`)

### Running the Producer

```sh
cd cmd/producer
 go run main.go
```

This will continuously produce random order messages to the `orders` stream.

### Running the Consumer

```sh
cd cmd/consumer
 go run main.go
```

This will consume and process messages from the `orders` stream as part of the `order_group` consumer group.

## How It Works

- **Producer**: Generates random orders and sends them to the Redis stream `orders`.
- **Consumer**: Reads messages from the `orders` stream, processes them, and acknowledges successful processing. Failed messages are retried up to 3 times, then sent to a dead-letter queue (`orders-dlq`).

## Example Code

### Producer (cmd/producer/main.go)
```go
cfg := config.Load()
r := stream.NewRedisStream(cfg.RedisAddr, "orders", "order_group")
ctx := context.Background()

for {
  o := order.Order{
    ID:     uuid.NewString(),
    Item:   "Product-X",
    Amount: rand.Intn(5) + 1,
  }
  err := r.Produce(ctx, o)
  // ...
}
```

### Consumer (cmd/consumer/main.go)
```go
cfg := config.Load()
r := stream.NewRedisStream(cfg.RedisAddr, "orders", "order_group")

r.Consume(context.Background(), "consumer-1", order.HandleOrder)
```

### Order Handler (internal/order/handler.go)
```go
func HandleOrder(o Order) error {
  fmt.Printf("[Handler] Processing Order: %+v\n", o)
  time.Sleep(1 * time.Second)
  return nil
}
```

## Observability

The project uses [OpenTelemetry](https://opentelemetry.io/) for basic tracing of produce and consume operations.

## Testing

Unit tests for the stream logic are provided in `internal/stream/stream_test.go` and use [testcontainers-go](https://github.com/testcontainers/testcontainers-go) to spin up a real Redis instance for integration testing.

Run tests:

```sh
go test ./internal/stream
```

## License

This project is for educational purposes. Please add a LICENSE file if you intend to use it in production.

## References
- [Redis Streams Documentation](https://redis.io/docs/data-types/streams/)
- [go-redis/redis](https://github.com/go-redis/redis)
- [OpenTelemetry for Go](https://opentelemetry.io/docs/instrumentation/go/)
