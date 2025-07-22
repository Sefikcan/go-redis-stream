package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go-redis-stream/config"
	"go-redis-stream/internal/order"
	"go-redis-stream/internal/stream"
	"math/rand"
	"time"
)

func main() {
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
		if err != nil {
			fmt.Println("Produce error:", err)
		} else {
			fmt.Println("Produced:", o)
		}
		time.Sleep(2 * time.Second)
	}
}
