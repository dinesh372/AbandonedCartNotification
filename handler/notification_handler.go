package handler

import (
	"context"
	"fmt"
	"quince/controller"
	"quince/entity"
)

type NotificationHandler interface {
	HandleCartEvent(ctx context.Context, event entity.CartEvent) error
}

func NewNotificationHandler(controller controller.NotificationController) NotificationHandler {
	return &handler{
		notificationController: controller,
	}
}

type handler struct {
	notificationController controller.NotificationController
}

func (h *handler) HandleCartEvent(ctx context.Context, event entity.CartEvent) error {
	fmt.Printf("Cart updated event API request received by notification service. Event = %+v\n", event)
	return h.notificationController.HandleCartEvent(ctx, event)
}
