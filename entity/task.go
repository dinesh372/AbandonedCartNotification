package entity

type TaskResultStatus string

const (
	SUCCESS TaskResultStatus = "success"
	FAILURE TaskResultStatus = "failure"
)

// Task is a unit of abandoned-cart notification work claimed from the schedule set.
type Task struct {
	CartID      string
	CartVersion int
	// This ensures we only delete a lease if it hasn't been modified since task
	// was returned.
	LeaseExpiresAt int
	ContentId      string
	IdempotencyKey string
}

type AcceptResultRequest struct {
	Task   Task
	Result TaskResult
}

type TaskResult struct {
	Status TaskResultStatus
}
