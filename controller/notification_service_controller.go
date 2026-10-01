package controller

import (
	"context"
	"fmt"
	"quince/entity"
	"quince/repository"
	"runtime/debug"
	"strconv"
	"sync"
	"time"
)

type NotificationController interface {
	HandleCartEvent(ctx context.Context, event entity.CartEvent) error
	GetTask(ctx context.Context) (*entity.Task, error)
	AcceptResult(ctx context.Context, request entity.AcceptResultRequest) error
}

/*
*
Using Mutex to simulate Redis single threadedness
*
*/
type notificationController struct {
	cartRepository                  repository.CartRepository
	notificationRepository          repository.NotificationRepository
	schedules                       repository.SortedSet
	leases                          repository.SortedSet
	abandonedCartNotificationConfig entity.AbandonedCartNotificationConfig
	experimentController            ExperimentController
	mutex                           sync.Mutex
}

func NewNotificationController(
	cartRepository repository.CartRepository,
	notificationRepository repository.NotificationRepository,
	schedules repository.SortedSet,
	leases repository.SortedSet,
	experimentController ExperimentController,
	abandonedCartNotificationConfig entity.AbandonedCartNotificationConfig,
) NotificationController {
	return &notificationController{
		cartRepository:                  cartRepository,
		notificationRepository:          notificationRepository,
		schedules:                       schedules,
		leases:                          leases,
		experimentController:            experimentController,
		abandonedCartNotificationConfig: abandonedCartNotificationConfig,
		mutex:                           sync.Mutex{},
	}
}

/*
*
HandleCartEvent - contains the business logic for handling a new cart event.
It performs following operations:
1. Create / Update / delete Notification based on current cart state.
2. Update / create schedules.
3. Delete leases if exists.

Note that the logic is Idempotent and thus safe to retry. The idempotency is achieved via
below decisions:
 1. A partial failure in a run won't block the progress on subsequent retry. For example on
    cart purchase event schedules and leases are deleted but notifications aren't due to timeouts.
    The subsequent retry would not try find schedules/leases to delete but would delete notifications.
 2. We always fetch the latest state of cart from CartService it helps ensure we can deal with out of
    order events. For example a Cart version updates from 0 --> 1 --> 2 & kafka event for version 1 is
    processed after cart is already updated to 2 then we would operate on version 2 of cart. This means
    we are relying on processing time and not event time.
 3. Two events for the same cart can't be processed in parallel due to topic being partioned by cartId.
    This lack of interleaving greatly simplifies the solution.

*
*/
func (n *notificationController) HandleCartEvent(ctx context.Context, event entity.CartEvent) error {
	n.mutex.Lock()
	defer n.mutex.Unlock()

	logPrefix := fmt.Sprintf(
		"notificationController::HandlerCartEvent : cartId = %s eventType = %s version = %d. ",
		event.CartID,
		event.EventType,
		event.Version,
	)

	cartID := event.CartID
	cart, err := n.cartRepository.GetCartByID(ctx, cartID)
	if err != nil {
		fmt.Println(logPrefix+"Failed to fetch cart details. error = ", err)
		return nil
	}

	// Regardless of cart update event we will delete schedules and leases. Schedules
	// might be recreated with different scheduled time.
	// If the schedule doesn't exist then Delete operation won't return an error
	// from Repository. As a future improvement we can start returning error from
	// repository while also allowing classfication of errors so that we can ignore
	// not found error here.
	err = n.RemoveSchedulesAndLeasesByLuaScript(cartID, logPrefix)
	if err != nil {
		return err
	}

	// In case we dial down the Experiment we need to make sure we delete schedules
	// & leases. Thats why we are removing them as well as Notifiation in the database.
	experiment := n.experimentController.GetExperimentGroup(cart.UserID)
	if cart == nil ||
		cart.Status != entity.CartStatusActive ||
		experiment.Result.Group != entity.Treatment {
		fmt.Printf(logPrefix+"Deleting notification for cart %+v experiment %+v\n", cart, experiment.Result)
		err = n.notificationRepository.DeleteNotification(ctx, cartID)
		return err
	}

	existingNotification, err := n.notificationRepository.GetNotificationByID(ctx, cartID)
	notificationCountTotal := 0
	if existingNotification != nil {
		notificationCountTotal = existingNotification.NotificationCountTotal
		fmt.Printf(logPrefix+"Updating existing version = %d new version = %d\n",
			existingNotification.Version,
			cart.Version,
		)
	}
	notification := entity.Notification{
		CartID:                             cartID,
		Version:                            cart.Version,
		NotificationCountForCurrentVersion: 0,
		// Keep the last value. Only a notification sent (email etc) can increment it
		NotificationCountTotal: notificationCountTotal,
	}

	if err = n.notificationRepository.UpsertNotification(ctx, notification); err != nil {
		fmt.Printf(logPrefix+"Failed to update notification. error = %v \n", err)
		return err
	}

	nextScheduleAt, _ := n.abandonedCartNotificationConfig.NotifyAt(time.Now(), 0)
	fmt.Printf(logPrefix+"Creating a new schedule at %v currentTime = %v difference = %v\n", nextScheduleAt, time.Now(), time.Until(nextScheduleAt))

	if err = n.schedules.Store(nextScheduleAt.Unix(), cartID); err != nil {
		fmt.Printf(logPrefix+"failed to create a new schedule. error = %v\n", err)
		return err
	}

	return nil
}

func (n *notificationController) RemoveSchedulesAndLeasesByLuaScript(cartID string, logPrefix string) error {

	scheduledAt, err := n.schedules.GetScore(cartID)
	if err != nil {
		return err
	}
	leaseExpiresAt, err := n.leases.GetScore(cartID)
	if err != nil {
		return err
	}
	if scheduledAt == nil && leaseExpiresAt == nil {
		fmt.Printf(logPrefix+"No schedules or leases were present for cart %s. \n", cartID)
		return nil
	}

	//fmt.Println(logPrefix+"Removing schedule which was scheduled at ", time.Unix(*scheduledAt, 0))
	err = n.schedules.Delete(cartID)
	if err != nil {
		return err
	}

	if leaseExpiresAt == nil {
		return nil
	}
	fmt.Println("Removing lease which expires at = ", *leaseExpiresAt)
	err = n.leases.Delete(cartID)
	return err
}

/*
  - GetTask - returns a new notification task.
    For durability and avoiding missed notifications we would execute dequeue from schedules
    and enqueue in leases using a Redis Lua Script. A Lua Script is sent as a single command
    to Redis server while it is shown as 2 API calls below. A lua script can't partially fail
    unless there are bugs in the script. Once a request reaches the Redis server it will execute
    to completion even if client has timed out.

    Note that in case of timeout we can't assume Redis has received the request so the request
    status would be treated as UNKOWN. We will have to retry which is a safe and Idempotent operation.

    For cases where GetTask succeeds but the worker fails then leases would be used to push
    the failed schedule back to schedules. The lease would be deleted when the worker submits
    the result back to notification service.

*
*/
func (n *notificationController) GetTask(ctx context.Context) (*entity.Task, error) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Successfully caught a panic: %v\n", r)
			debug.PrintStack()
		}
	}()
	now := time.Now()
	entry, err := n.schedules.Dequeue(now.Unix())
	if err != nil {
		fmt.Printf("notificationController::GetTask : failed to dequeue schedule. error = %v\n", err)
		return nil, err
	}
	if entry == nil {
		return nil, nil
	}

	leaseExpiresAt := now.Add(n.abandonedCartNotificationConfig.LeaseDurationOrDefault())
	if err = n.leases.Store(leaseExpiresAt.Unix(), entry.Key()); err != nil {
		fmt.Printf("notificationController::GetTask : failed to create lease for cartId = %s. error = %v\n", entry.Key(), err)
		return nil, err
	}

	notification, err := n.notificationRepository.GetNotificationByID(ctx, entry.Key())
	if err != nil {
		return nil, err
	}

	cart, err := n.cartRepository.GetCartByID(ctx, entry.Key())
	if err != nil {
		return nil, err
	}

	// Check just before sending notification that Cart is still Active.
	// TODO: Also check for when cart was last updated/version to handle cases when notification
	// is for an older cart version.
	if notification == nil || cart.Status != entity.CartStatusActive {
		fmt.Printf("notificationController::GetTask : notification doesn't exist for cartId = %s.\n", notification.CartID)
		n.RemoveSchedulesAndLeasesByLuaScript(entry.Key(), "notificationController::GetTask")
		return nil, nil
	}

	idempotentKey := entry.Key() + "_" + strconv.Itoa(notification.NotificationCountTotal)

	task := &entity.Task{
		CartID:         entry.Key(),
		CartVersion:    cart.Version,
		IdempotencyKey: idempotentKey,
		LeaseExpiresAt: int(leaseExpiresAt.Unix()),
		ContentId:      notification.ContentId,
	}
	fmt.Printf("notificationController::GetTask : claimed cartId = %s leaseUntil = %v\n",
		task.CartID, task.LeaseExpiresAt)
	return task, nil
}

func (n *notificationController) AcceptResult(
	ctx context.Context,
	request entity.AcceptResultRequest) error {

	notification, err := n.notificationRepository.GetNotificationByID(ctx, request.Task.CartID)
	if err != nil {
		return err
	}

	cart, err := n.cartRepository.GetCartByID(ctx, request.Task.CartID)
	if err != nil {
		return err
	}

	// We should ideally do a conditional write here to prevent deleting lease
	// created by another cart update. This is a race condition.

	// Another Event modified the cart delete the lease if lease time matches matches.
	if request.Task.CartVersion != cart.Version {
		n.leases.ConditionalDelete(request.Task.CartID, int64(request.Task.LeaseExpiresAt))
		return nil
	}

	nextScheduleAt, shouldSchedule := n.abandonedCartNotificationConfig.NotifyAt(time.Now(), notification.NotificationCountForCurrentVersion)
	if shouldSchedule {
		difference := nextScheduleAt.Sub(time.Now()).Milliseconds()
		fmt.Printf("Scheduling another notification for the same cart and version. currenttime = %+v scheduledAt = %+v count = %+v difference = %v \n", time.Now(), nextScheduleAt, notification.NotificationCountForCurrentVersion, difference)
		n.schedules.Store(nextScheduleAt.Unix(), request.Task.CartID)
	}

	notification.NotificationCountForCurrentVersion += 1
	notification.NotificationCountTotal += 1
	if err = n.notificationRepository.UpdateNotification(ctx, notification); err != nil {
		return err
	}

	// Deleting lease at the end so that we can recover from intermittent failures.
	err = n.leases.ConditionalDelete(request.Task.CartID, int64(request.Task.LeaseExpiresAt))

	return err
}
