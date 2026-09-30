package repository

import (
	"context"
	"fmt"
)

type dummyDatabase[T any] struct {
	data map[string]T
}

func NewDummyDatabase[T any]() *dummyDatabase[T] {
	return &dummyDatabase[T]{
		data: make(map[string]T),
	}
}

func (d *dummyDatabase[T]) Create(ctx context.Context, key string, value T) error {

	if _, ok := d.data[key]; ok {
		return fmt.Errorf("key already exists")
	}
	d.data[key] = value
	return nil
}

func (d *dummyDatabase[T]) Get(ctx context.Context, key string) *T {
	value, ok := d.data[key]
	if !ok {
		return nil
	}
	return &value
}

func (d *dummyDatabase[T]) Update(ctx context.Context, key string, value T) error {
	if _, ok := d.data[key]; !ok {
		return fmt.Errorf("key not found")
	}
	d.data[key] = value
	return nil
}

func (d *dummyDatabase[T]) Upsert(ctx context.Context, key string, value T) error {
	d.data[key] = value
	return nil
}

func (d *dummyDatabase[T]) Delete(ctx context.Context, key string) error {
	delete(d.data, key)
	return nil
}

/*
*
Clear sets the value for a key to an empty value.
*
*/
func (d *dummyDatabase[T]) Clear(ctx context.Context, key string) error {
	if _, ok := d.data[key]; !ok {
		return fmt.Errorf("key not found")
	}
	var empty T
	d.data[key] = empty
	return nil
}
