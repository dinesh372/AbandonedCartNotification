package entity

type CartEventType string

const (
	CartCreated   CartEventType = "cart_created"
	CartUpdated   CartEventType = "cart_updated"
	CartCleared   CartEventType = "cart_cleared"
	CartCompleted CartEventType = "cart_completed"
)

type CartEvent struct {
	CartID    string        `json:"cart_id"`
	EventType CartEventType `json:"event_type"`
	Version   int           `json:"version"`
}
