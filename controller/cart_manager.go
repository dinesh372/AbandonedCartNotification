package controller

import (
	"context"
	"fmt"
	"quince/entity"
	"quince/repository"
	"time"
)

type CartController interface {
	CreateCart(ctx context.Context, cart_id string, user_id string) error
	AddItemToCart(ctx context.Context, cart_id string, item_id string, quantity int, productId string) error
	CompleteCart(ctx context.Context, cart_id string) error
	GetCartUpdatesQueue() chan entity.CartEvent
}

type cartController struct {
	CartRepository repository.CartRepository
	// A buffered channel to send cart updates to the client.
	// A buffered channel means we don't wait for the client to read the message.
	CartUpdatesQueue chan entity.CartEvent
}

/*
NewCartController creates a new cart controller.
Ideally Repository should be injected via dependency injection.
*/
func NewCartController(repository repository.CartRepository, queue chan entity.CartEvent) CartController {
	return &cartController{
		CartRepository:   repository,
		CartUpdatesQueue: queue,
	}
}

func (c *cartController) GetCartUpdatesQueue() chan entity.CartEvent {
	return c.CartUpdatesQueue
}

func GetDummyItems() []entity.CartItem {
	return []entity.CartItem{
		{
			ID:        "1",
			Quantity:  2,
			ProductId: "1",
		},
		{
			ID:        "2",
			Quantity:  1,
			ProductId: "2",
		},
	}
}

func (c *cartController) CreateCart(ctx context.Context, cartID string, userID string) error {

	fmt.Println("CartController : Creating a new cart. cart_id = ", cartID, " user_id = ", userID)

	// Using Dummy Items as we don't care about the items for now.
	cart := &entity.Cart{
		ID:        cartID,
		UserID:    userID,
		Items:     GetDummyItems(),
		Status:    entity.CartStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := c.CartRepository.CreateCart(ctx, cart); err != nil {
		return err
	}

	fmt.Println("CartController : Successfully created a new cart. cart_id = ", cartID, " user_id = ", userID)

	c.CartUpdatesQueue <- entity.CartEvent{
		CartID:    cart.ID,
		EventType: entity.CartCreated,
		Version:   cart.Version,
	}

	return nil
}

func (c *cartController) AddItemToCart(ctx context.Context, cartID string, itemID string, quantity int, productId string) error {
	fmt.Println("CartController::AddItem : Adding a new item to cart. cart_id = ", cartID, " item_id = ", itemID)

	cart, err := c.CartRepository.GetCartByID(ctx, cartID)
	if err != nil {
		fmt.Println("CartController::AddItem : Adding a new item failed as no cart found for cart_id = ", cartID, ". error = ", err)
		return err
	}

	cart.UpdatedAt = time.Now()
	cart.Version = cart.Version + 1
	cart.Items = append(cart.Items, entity.CartItem{
		ID:        itemID,
		Quantity:  quantity,
		ProductId: productId,
	})

	if err := c.CartRepository.UpdateCart(ctx, cart); err != nil {
		fmt.Println("CartController::AddItem : Adding a new item failed  cart_id = ", cartID, ". error = ", err)
		return fmt.Errorf("Failed to update cart. %s", err.Error())
	}

	fmt.Println("CartController::AddItem : Successfully added a new item to cart  cart_id = ", cartID, ". item_id ", itemID)

	c.CartUpdatesQueue <- entity.CartEvent{
		CartID:    cart.ID,
		Version:   cart.Version,
		EventType: entity.CartUpdated,
	}

	return nil
}

func (c *cartController) CompleteCart(ctx context.Context, cartID string) error {
	fmt.Println("CartController::CompleteCart : Completing cart. cart_id = ", cartID)

	cart, err := c.CartRepository.GetCartByID(ctx, cartID)
	if err != nil {
		fmt.Println("CartController::CompleteCart : failed to fetch cart details. cart_id = ", cartID, " error = ", err)
		return fmt.Errorf("Failed to mark cart as complete. %s", err.Error())
	}

	cart.Status = entity.CartStatusCompleted
	cart.UpdatedAt = time.Now()
	cart.Version = cart.Version + 1

	if err := c.CartRepository.UpdateCart(ctx, cart); err != nil {
		fmt.Println("CartController::CompleteCart : failed to update cart details. cart_id = ", cartID, " error = ", err)
		return err
	}

	fmt.Println("CartController::CompleteCart : Successfully marked the cart as complete. cart_id = ", cartID)

	c.CartUpdatesQueue <- entity.CartEvent{
		CartID:    cart.ID,
		Version:   cart.Version,
		EventType: entity.CartCompleted,
	}
	return nil
}
