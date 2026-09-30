package consumer

import (
	"context"
	"fmt"
	"quince/entity"
	"quince/handler"
)

/**
* CartUpdatesConsumer - handles messages from Cart Updates kafka queue.
* The queue is modeled as a buffered go channel
**/
type CartUpdatesConsumer interface {
	StartConsumingEvents()
}

type cartUpdatesConsumer struct {
	cartUpdatesQueue    chan entity.CartEvent
	notificationService handler.NotificationHandler
}

func NewCartUpdatesConsumer(
	notifcationService handler.NotificationHandler,
	cartUpdatesQueue chan entity.CartEvent,
) CartUpdatesConsumer {
	return &cartUpdatesConsumer{
		cartUpdatesQueue:    cartUpdatesQueue,
		notificationService: notifcationService,
	}
}

/**
* Start a go-routine(user space thread) to handle cart update events.
* This go-routine will continue to execute in the background while the
* parent function will return immediately.
* The go-routine would call Notification service whenever a cart update
* event is received.
*
* For simplicity we can assume go rountine as similar to threads in
* other programming languages. Virtual thread in Java introduced as
* part of project loom is a good approximation
* https://tinyurl.com/32fprzkm
*
* NOTE that we don't intend to simulate multi partitions of kafka here &
* all messages would be read in a sequence from the go channel.
**/
func (c *cartUpdatesConsumer) StartConsumingEvents() {
	go func() {
		for event := range c.cartUpdatesQueue {
			fmt.Printf("Cart updated event received by consumer. Event = %+v\n", event)
			ctx := context.Background()

			c.notificationService.HandleCartEvent(ctx, event)
			// Add Retry
		}
	}()
}
