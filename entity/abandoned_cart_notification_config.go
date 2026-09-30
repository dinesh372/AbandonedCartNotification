package entity

import (
	"time"
)

// AbandonedCartNotificationConfig defines the timing ladder for reminding customers
// about carts that have gone idle.
type AbandonedCartNotificationConfig struct {
	Steps []AbandonedCartNotificationStep `json:"steps"`
	// LeaseDuration is how long a worker holds a dequeued schedule before another
	// worker may claim it. Defaults to 5 minutes when unset.
	LeaseDuration time.Duration `json:"lease_duration"`
}

// AbandonedCartNotificationStep is one reminder in the ladder.
type AbandonedCartNotificationStep struct {
	WaitFromLastCartActivity time.Duration `json:"wait_from_last_cart_activity"`
}

const DefaultLeaseDuration = 5 * time.Minute

/**
func DefaultAbandonedCartNotificationConfig() AbandonedCartNotificationConfig {
	return AbandonedCartNotificationConfig{
		Steps: []AbandonedCartNotificationStep{
			{WaitFromLastCartActivity: 30 * time.Minute},
			{WaitFromLastCartActivity: 12 * time.Hour},
			{WaitFromLastCartActivity: 24 * time.Hour},
		},
		LeaseDuration: DefaultLeaseDuration,
	}
} **/

// Using shorter notification window for testing.
// Values as per the spec are above.
func DefaultAbandonedCartNotificationConfig() AbandonedCartNotificationConfig {
	return AbandonedCartNotificationConfig{
		Steps: []AbandonedCartNotificationStep{
			{WaitFromLastCartActivity: 1 * time.Second},
			{WaitFromLastCartActivity: 5 * time.Second},
			{WaitFromLastCartActivity: 10 * time.Second},
		},
		LeaseDuration: DefaultLeaseDuration,
	}
}

func (c AbandonedCartNotificationConfig) LeaseDurationOrDefault() time.Duration {
	if c.LeaseDuration <= 0 {
		return DefaultLeaseDuration
	}
	return c.LeaseDuration
}

func (c AbandonedCartNotificationConfig) StepCount() int {
	return len(c.Steps)
}

// NotifyAt returns the absolute time when the step at stepIndex should be sent,
// measured from lastCartActivity. stepIndex is zero-based.
func (c AbandonedCartNotificationConfig) NotifyAt(lastCartActivity time.Time, stepIndex int) (time.Time, bool) {
	if stepIndex < 0 || stepIndex >= len(c.Steps) {
		return time.Time{}, false
	}
	return lastCartActivity.Add(c.Steps[stepIndex].WaitFromLastCartActivity), true
}
