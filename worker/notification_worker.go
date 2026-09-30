package worker

import (
	"context"
	"fmt"
	"quince/controller"
	"quince/entity"
	"runtime/debug"
	"time"
)

/**
* NotficationWorker - Polls the notification service for notifcations to be sent
**/
type NotifcationWorker interface {
	Start()
}

type notifcationWorker struct {
	cartUpdatesQueue chan entity.CartEvent
	// Using controller as the entry point instead of handler to save
	// some boiler plate
	notificationService controller.NotificationController
	notifier            controller.CustomerNotifier
}

func NewNotificationWorker(
	notifcationService controller.NotificationController,
	notifier controller.CustomerNotifier,
) NotifcationWorker {
	return &notifcationWorker{
		notificationService: notifcationService,
		notifier:            notifier,
	}
}

/**
* Start a go-routine(user space thread) to poll messages from notifcationService.
**/
func (w *notifcationWorker) Start() {
	go func() {
		if r := recover(); r != nil {
			fmt.Printf("Successfully caught a panic in worker: %v\n", r)
			debug.PrintStack()
		}
		for {
			ctx := context.Background()

			task, err := w.notificationService.GetTask(ctx)
			// Even in case of errors like timeouts this is safe due to
			// leases.
			if err != nil || task == nil {
				continue
			}

			fmt.Printf("NotificationWorker : processing task %v currentTime = %v\n", task, time.Now())

			// When the task is delivered again due to lease the idempotency key
			// ensures we don't notify customer multiple times when its unknown whether
			// the customer was notified or not e.g. timeouts.
			err = w.notifier.NotifyCustomer(*task, task.IdempotencyKey)
			if err != nil {
				continue
			}

			acceptRequest := entity.AcceptResultRequest{
				Task: *task,
				Result: entity.TaskResult{
					Status: entity.SUCCESS,
				},
			}
			w.notificationService.AcceptResult(ctx, acceptRequest)
		}
	}()
}
