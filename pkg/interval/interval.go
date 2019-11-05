package interval

import (
	"context"
	"time"
)

// Ticker is a simple interface which defines how to implement
// interval
type Ticker interface {
	Tick(ctx context.Context) error
}

// Run calls ticker.Tick method based on duration
// The given context can be canceled to stop the Run
// NOTE: this method is blocking
func Run(ctx context.Context, action Ticker, duration time.Duration) error {
	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			err := action.Tick(ctx)
			if err != nil {
				return err
			}

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
