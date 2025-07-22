package stream

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go-redis-stream/internal/order"
	"testing"
	"time"
)

func TestProduceConsume(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "redis:6.2",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	redisC, _ := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	defer redisC.Terminate(ctx)

	host, _ := redisC.Host(ctx)
	port, _ := redisC.MappedPort(ctx, "6379")
	addr := host + ":" + port.Port()

	rs := NewRedisStream(addr, "test-orders", "test-group")
	orderSent := order.Order{ID: "abc", Item: "Test", Amount: 10}

	err := rs.Produce(ctx, orderSent)
	assert.NoError(t, err)

	done := make(chan bool)
	go rs.Consume(ctx, "test-consumer", func(o order.Order) error {
		assert.Equal(t, orderSent.ID, o.ID)
		done <- true
		return nil
	})

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}
