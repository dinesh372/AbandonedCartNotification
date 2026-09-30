package repository

import (
	"context"
	"quince/entity"
)

type CartRepository interface {
	CreateCart(ctx context.Context, cart *entity.Cart) error
	GetCartByID(ctx context.Context, id string) (*entity.Cart, error)
	UpdateCart(ctx context.Context, cart *entity.Cart) error
	ClearCart(ctx context.Context, id string) error
}

type cartRepository struct {
	db *dummyDatabase[entity.Cart]
}

func NewCartRepository() CartRepository {
	return &cartRepository{
		db: NewDummyDatabase[entity.Cart](),
	}
}

func (r *cartRepository) CreateCart(ctx context.Context, cart *entity.Cart) error {
	return r.db.Create(ctx, cart.ID, *cart)
}

func (r *cartRepository) GetCartByID(ctx context.Context, id string) (*entity.Cart, error) {
	return r.db.Get(ctx, id), nil
}

func (r *cartRepository) UpdateCart(ctx context.Context, cart *entity.Cart) error {
	return r.db.Update(ctx, cart.ID, *cart)
}

func (r *cartRepository) ClearCart(ctx context.Context, id string) error {
	return r.db.Clear(ctx, id)
}
