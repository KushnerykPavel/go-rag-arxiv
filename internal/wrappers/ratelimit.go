package wrappers

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter limits calls to at most rps per second.
// Wrap any call with Do to enforce the limit.
type RateLimiter struct {
	limiter *rate.Limiter
}

// NewRateLimiter returns a RateLimiter that allows at most rps calls per second.
// rps must be greater than zero.
func NewRateLimiter(rps int) (*RateLimiter, error) {
	if rps <= 0 {
		return nil, fmt.Errorf("rps must be greater than zero, got %d", rps)
	}
	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Every(time.Duration(rps)*time.Minute), rps),
	}, nil
}

// Do waits until a token is available, then calls fn.
// Returns ctx.Err() if the context is cancelled while waiting.
func (r *RateLimiter) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	if err := r.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("rate limiter: %w", err)
	}
	return fn(ctx)
}
