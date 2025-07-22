package main

import (
	"context"
	"go-redis-stream/config"
	"go-redis-stream/internal/order"
	"go-redis-stream/internal/stream"
)

func main() {
	cfg := config.Load()
	r := stream.NewRedisStream(cfg.RedisAddr, "orders", "order_group")

	r.Consume(context.Background(), "consumer-1", order.HandleOrder)
}
