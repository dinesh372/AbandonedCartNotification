package main

import (
	"context"
	"fmt"
	"quince/consumer"
	"quince/controller"
	"quince/entity"
	"quince/handler"
	"quince/repository"
	"time"
)

func main() {
	fmt.Println("Hello, World!")

	cartRepository := repository.NewCartRepository()
	cartUpdatesQueue := make(chan entity.CartEvent, 100)
	cartController := controller.NewCartController(cartRepository, cartUpdatesQueue)
	//notifcationController := controller.NewNotificationController()
	notificationService := handler.NewNotificationHandler(nil)

	cartUpdatesConsumer := consumer.NewCartUpdatesConsumer(notificationService, cartUpdatesQueue)
	cartUpdatesConsumer.StartConsumingEvents()

	cartController.CreateCart(context.Background(), "1", "1")
	cartController.AddItemToCart(context.Background(), "1", "3", 1, "3")
	cartController.CompleteCart(context.Background(), "1")

	time.Sleep(1 * time.Second)

}
