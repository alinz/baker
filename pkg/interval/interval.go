package interval

import (
	"context"
	"time"
)

type Ticker interface {
	Tick(ctx context.Context) error
}

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
