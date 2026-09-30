package controller

import (
	"fmt"
	"quince/entity"
)

type CustomerNotifier interface {
	NotifyCustomer(task entity.Task, idempotencyKey string) error
}

type notifier struct {
}

func NewCustomerNotifier() CustomerNotifier {
	return &notifier{}
}

func (n *notifier) NotifyCustomer(task entity.Task, idempotencyKey string) error {
	fmt.Printf("Notify customer with contentId = %s IdempotencyKey = %s task = %+v\n", task.ContentId, idempotencyKey, task)
	return nil
}
