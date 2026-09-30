package repository

import (
	"context"
	"quince/entity"
)

type NotificationRepository interface {
	CreateNotification(ctx context.Context, notification entity.Notification) error
	GetNotificationByID(ctx context.Context, id string) (*entity.Notification, error)
	UpdateNotification(ctx context.Context, notification *entity.Notification) error
	UpsertNotification(ctx context.Context, notification entity.Notification) error
	DeleteNotification(ctx context.Context, id string) error
}

type notificationRepository struct {
	db *dummyDatabase[entity.Notification]
}

func NewNotificationRepository() NotificationRepository {
	return &notificationRepository{
		db: NewDummyDatabase[entity.Notification](),
	}
}

func (r *notificationRepository) CreateNotification(ctx context.Context, notification entity.Notification) error {
	return r.db.Create(ctx, notification.CartID, notification)
}

func (r *notificationRepository) GetNotificationByID(ctx context.Context, id string) (*entity.Notification, error) {
	return r.db.Get(ctx, id), nil
}

func (r *notificationRepository) UpdateNotification(ctx context.Context, notification *entity.Notification) error {
	return r.db.Update(ctx, notification.CartID, *notification)
}

func (r *notificationRepository) UpsertNotification(ctx context.Context, notification entity.Notification) error {
	return r.db.Upsert(ctx, notification.CartID, notification)
}

func (r *notificationRepository) DeleteNotification(ctx context.Context, id string) error {
	return r.db.Delete(ctx, id)
}
