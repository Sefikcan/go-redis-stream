package order

import (
	"fmt"
	"time"
)

func HandleOrder(o Order) error {
	fmt.Printf("[Handler] Processing Order: %+v\n", o)
	time.Sleep(1 * time.Second)
	return nil
}
