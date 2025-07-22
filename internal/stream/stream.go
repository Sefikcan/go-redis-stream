package stream

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-redis/redis/v8"
	"go-redis-stream/internal/order"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"log"
	"time"
)

type RedisStream struct {
	Client redis.UniversalClient
	Stream string
	Group  string
	Tracer trace.Tracer
}

func NewRedisStream(addr, stream, group string) *RedisStream {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	tracer := otel.Tracer("stream")

	return &RedisStream{
		Client: client,
		Stream: stream,
		Group:  group,
		Tracer: tracer,
	}
}

func (r *RedisStream) Produce(ctx context.Context, order order.Order) error {
	ctx, span := r.Tracer.Start(ctx, "Produce")
	defer span.End()

	data, _ := json.Marshal(order)
	_, err := r.Client.XAdd(ctx, &redis.XAddArgs{
		Stream: r.Stream,
		Values: map[string]interface{}{"data": data},
	}).Result()

	return err
}

func (r *RedisStream) Consume(ctx context.Context, consumeName string, handler func(order.Order) error) {
	_ = r.Client.XGroupCreateMkStream(ctx, r.Stream, r.Group, "0").Err()

	for {
		ctx, span := r.Tracer.Start(ctx, "Consume")
		streams, err := r.Client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    r.Group,
			Consumer: consumeName,
			Streams:  []string{r.Stream, ">"},
			Block:    5 * time.Second,
		}).Result()
		span.End()
		if err != nil && !errors.Is(err, redis.Nil) {
			log.Printf("[Consumer] ReadGroup error: %v", err)
			continue
		}

		for _, stream := range streams {
			for _, message := range stream.Messages {
				raw := message.Values["data"].(string)
				var o order.Order
				_ = json.Unmarshal([]byte(raw), &o)

				success := false
				for i := 0; i < 3; i++ {
					if err := handler(o); err == nil {
						success = true
						break
					}
					time.Sleep(time.Second)
				}

				if success {
					r.Client.XAck(ctx, r.Stream, r.Group, message.ID)
				} else {
					log.Printf("[DLQ] Failed processing message ID: %s", message.ID)
					r.Client.XAdd(ctx, &redis.XAddArgs{
						Stream: r.Stream + "-dlq",
						Values: map[string]interface{}{"data": raw},
					})
					r.Client.XAck(ctx, r.Stream, r.Group, message.ID)
				}
			}
		}
	}
}
