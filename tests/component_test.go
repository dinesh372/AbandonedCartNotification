package tests

import (
	"context"
	"quince/consumer"
	"quince/controller"
	"quince/entity"
	"quince/handler"
	"quince/repository"
	"quince/worker"
	"testing"
	"time"
)

type DependencyContainer struct {
	// Notification service endpoints handler.
	notificationService handler.NotificationHandler

	// Controller - contains the business logic.
	cartController         controller.CartController
	experimentController   controller.ExperimentController
	notificationController controller.NotificationController
	customerNotifer        controller.CustomerNotifier

	// Consumer - cart updates Queue and Consumer
	cartUpdatesQueue    chan entity.CartEvent
	cartUpdatesConsumer consumer.CartUpdatesConsumer

	// Worker - polls for tasks and send notifications to the customer.
	notificationWorker worker.NotifcationWorker

	// Repository - for carts, notifcations, schedules and leases
	cartRepository         repository.CartRepository
	notificationRepository repository.NotificationRepository
	schedules              repository.SortedSet
	leases                 repository.SortedSet
}

func createDependencies() DependencyContainer {
	cartRepository := repository.NewCartRepository()
	notificationRepository := repository.NewNotificationRepository()
	schedules := repository.NewSortedSet("schedules")
	leases := repository.NewSortedSet("leases")

	experimentController := controller.NewExperimentController()
	notificationController := controller.NewNotificationController(
		cartRepository,
		notificationRepository,
		schedules,
		leases,
		experimentController,
		entity.DefaultAbandonedCartNotificationConfig(),
	)

	notificationService := handler.NewNotificationHandler(notificationController)

	cartUpdatesQueue := make(chan entity.CartEvent, 100)
	cartUpdatesConsumer := consumer.NewCartUpdatesConsumer(notificationService, cartUpdatesQueue)

	cartController := controller.NewCartController(cartRepository, cartUpdatesQueue)

	customerNotifier := controller.NewCustomerNotifier()
	notificationWorker := worker.NewNotificationWorker(notificationController, customerNotifier)

	cartUpdatesConsumer.StartConsumingEvents()
	notificationWorker.Start()

	return DependencyContainer{

		notificationService:    notificationService,
		cartController:         cartController,
		experimentController:   experimentController,
		notificationController: notificationController,
		cartUpdatesQueue:       cartUpdatesQueue,
		cartUpdatesConsumer:    cartUpdatesConsumer,

		notificationWorker: notificationWorker,

		cartRepository:         cartRepository,
		notificationRepository: notificationRepository,
		schedules:              schedules,
		leases:                 leases,
	}
}

func (d DependencyContainer) StartQueueConsumerAndNotificationPollers() {
	d.cartUpdatesConsumer.StartConsumingEvents()
	d.notificationWorker.Start()
}

func Test_Cart_Abandoned(t *testing.T) {
	container := createDependencies()
	container.StartQueueConsumerAndNotificationPollers()

	ctx := context.Background()

	container.cartController.CreateCart(ctx, "cart-1", "user-1")
	time.Sleep(10 * time.Millisecond)
	container.cartController.AddItemToCart(ctx, "cart-1", "itemid-1", 1, "productid-1")
	time.Sleep(10 * time.Millisecond)
	container.cartController.AddItemToCart(ctx, "cart-1", "itemid-2", 1, "productid-2")
	time.Sleep(25 * time.Second)
}

func Test_CartPurchased_BeforeAbandoned(t *testing.T) {
	container := createDependencies()
	container.StartQueueConsumerAndNotificationPollers()

	ctx := context.Background()
	container.cartController.CreateCart(ctx, "cart-1", "user-1")
	time.Sleep(10 * time.Millisecond)
	container.cartController.AddItemToCart(ctx, "cart-1", "itemid-1", 1, "productid-1")
	time.Sleep(10 * time.Millisecond)
	container.cartController.CompleteCart(ctx, "cart-1")
	time.Sleep(1 * time.Second)
}

func Test_UserInControlGroup_NotificationNotSent(t *testing.T) {
	container := createDependencies()
	container.StartQueueConsumerAndNotificationPollers()

	ctx := context.Background()
	container.cartController.CreateCart(ctx, "cart-1", "user-2")
	time.Sleep(10 * time.Millisecond)
	container.cartController.AddItemToCart(ctx, "cart-1", "itemid-1", 1, "productid-1")
	time.Sleep(10 * time.Millisecond)
	container.cartController.CompleteCart(ctx, "cart-1")
	time.Sleep(1 * time.Second)
}
