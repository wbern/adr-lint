// Package retry provides bounded exponential-backoff retries for
// transient failures (HTTP 429 / 5xx). Options.Rand and Options.Sleep
// are injectable so tests can drive backoff deterministically without
// real time passing.
package retry

import (
	"errors"
	"math"
	"math/rand"
	"time"
)

// RetryableError signals that the caller should retry after an
// optional, server-specified delay.
type RetryableError struct {
	Message      string
	StatusCode   int
	RetryAfterMs *int64
}

func (e *RetryableError) Error() string { return e.Message }

var retryableStatusCodes = map[int]struct{}{
	429: {}, 500: {}, 502: {}, 503: {}, 504: {},
}

// IsRetryableStatus reports whether an HTTP status code should trigger
// a retry (429 plus the standard 5xx subset).
func IsRetryableStatus(status int) bool {
	_, ok := retryableStatusCodes[status]
	return ok
}

// Options configures WithRetry. Sleep and Rand are injectable hooks
// for deterministic testing; nil falls back to time.Sleep / rand.Float64.
type Options struct {
	MaxRetries  *int
	BaseDelayMs *int64
	MaxDelayMs  *int64
	Sleep       func(time.Duration)
	Rand        func() float64
}

// WithRetry calls fn up to MaxRetries+1 times, retrying only when fn
// returns a *RetryableError.
func WithRetry[T any](fn func(attempt int) (T, error), opts Options) (T, error) {
	maxRetries := 5
	if opts.MaxRetries != nil {
		maxRetries = *opts.MaxRetries
	}
	baseDelayMs := int64(5000)
	if opts.BaseDelayMs != nil {
		baseDelayMs = *opts.BaseDelayMs
	}
	maxDelayMs := int64(60_000)
	if opts.MaxDelayMs != nil {
		maxDelayMs = *opts.MaxDelayMs
	}
	sleep := opts.Sleep
	if sleep == nil {
		sleep = time.Sleep
	}
	randFn := opts.Rand
	if randFn == nil {
		randFn = rand.Float64
	}

	var zero T
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		result, err := fn(attempt)
		if err == nil {
			return result, nil
		}
		var re *RetryableError
		if !errors.As(err, &re) || attempt == maxRetries {
			return zero, err
		}
		lastErr = err
		var delayMs int64
		if re.RetryAfterMs != nil {
			delayMs = *re.RetryAfterMs
			if delayMs > maxDelayMs {
				delayMs = maxDelayMs
			}
		} else {
			base := float64(baseDelayMs) * math.Pow(2, float64(attempt))
			jitter := base * 0.5 * randFn()
			delayMs = int64(base + jitter)
			if delayMs > maxDelayMs {
				delayMs = maxDelayMs
			}
		}
		sleep(time.Duration(delayMs) * time.Millisecond)
	}
	return zero, lastErr
}
