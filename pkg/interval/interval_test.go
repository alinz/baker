package interval_test

import (
	"context"
	"testing"
	"time"

	"github.com/alinz/baker/pkg/interval"
)

type DummyTickerFn func(ctx context.Context) error

func (dt DummyTickerFn) Tick(ctx context.Context) error {
	return dt(ctx)
}

var _ interval.Ticker = (*DummyTickerFn)(nil)

func TestInterval(t *testing.T) {
	count := 1
	dummyTicker := DummyTickerFn(func(cxt context.Context) error {
		count++
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := interval.Run(ctx, dummyTicker, 500*time.Millisecond)
	if err.Error() != "context deadline exceeded" {
		t.Fatalf("expected '%s' error but got this '%s'", "context deadline exceeded", err.Error())
	}

	if count != 5 {
		t.Fatal("count is not incremented", count)
	}
}
