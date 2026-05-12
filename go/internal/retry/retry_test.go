package retry

import (
	"errors"
	"testing"
	"time"
)

func TestWithRetry_ReturnsResultOnFirstSuccess(t *testing.T) {
	calls := 0
	fn := func(attempt int) (string, error) {
		calls++
		return "success", nil
	}

	got, err := WithRetry(fn, Options{Sleep: noSleep})
	if err != nil {
		t.Fatalf("WithRetry: %v", err)
	}
	if got != "success" {
		t.Errorf("got %q, want %q", got, "success")
	}
	if calls != 1 {
		t.Errorf("fn called %d times, want 1", calls)
	}
}

func TestWithRetry_RetriesOnRetryableErrorThenSucceeds(t *testing.T) {
	calls := 0
	fn := func(attempt int) (string, error) {
		calls++
		if calls == 1 {
			return "", &RetryableError{Message: "rate limited", StatusCode: 429}
		}
		return "success after retry", nil
	}

	got, err := WithRetry(fn, Options{BaseDelayMs: i64ptr(0), Sleep: noSleep})
	if err != nil {
		t.Fatalf("WithRetry: %v", err)
	}
	if got != "success after retry" {
		t.Errorf("got %q", got)
	}
	if calls != 2 {
		t.Errorf("fn called %d times, want 2", calls)
	}
}

func TestWithRetry_NonRetryableErrorReturnsImmediately(t *testing.T) {
	calls := 0
	want := errors.New("Permission denied")
	fn := func(attempt int) (string, error) {
		calls++
		return "", want
	}

	_, err := WithRetry(fn, Options{BaseDelayMs: i64ptr(0), Sleep: noSleep})
	if !errors.Is(err, want) {
		t.Errorf("got err %v, want %v", err, want)
	}
	if calls != 1 {
		t.Errorf("fn called %d times, want 1", calls)
	}
}

func TestWithRetry_ThrowsAfterMaxRetries(t *testing.T) {
	calls := 0
	fn := func(attempt int) (string, error) {
		calls++
		return "", &RetryableError{Message: "rate limited", StatusCode: 429}
	}

	maxRetries := 2
	_, err := WithRetry(fn, Options{MaxRetries: &maxRetries, BaseDelayMs: i64ptr(0), Sleep: noSleep})
	var re *RetryableError
	if !errors.As(err, &re) || re.Message != "rate limited" {
		t.Errorf("got err %v, want RetryableError(rate limited)", err)
	}
	if calls != 3 {
		t.Errorf("fn called %d times, want 3 (1 initial + 2 retries)", calls)
	}
}

func TestWithRetry_ExponentialBackoffDelays(t *testing.T) {
	var delays []time.Duration
	sleep := func(d time.Duration) { delays = append(delays, d) }
	zeroRand := func() float64 { return 0 }

	calls := 0
	fn := func(attempt int) (string, error) {
		calls++
		if calls <= 2 {
			return "", &RetryableError{Message: "rate limited", StatusCode: 429}
		}
		return "success", nil
	}

	_, err := WithRetry(fn, Options{BaseDelayMs: i64ptr(100), Sleep: sleep, Rand: zeroRand})
	if err != nil {
		t.Fatalf("WithRetry: %v", err)
	}
	// attempt 0 backoff: 100 * 2^0 = 100ms; attempt 1: 100 * 2^1 = 200ms
	if len(delays) != 2 {
		t.Fatalf("recorded %d delays, want 2", len(delays))
	}
	if delays[0] != 100*time.Millisecond {
		t.Errorf("delays[0] = %v, want 100ms", delays[0])
	}
	if delays[1] != 200*time.Millisecond {
		t.Errorf("delays[1] = %v, want 200ms", delays[1])
	}
}

func TestWithRetry_AddsJitter(t *testing.T) {
	var delays []time.Duration
	sleep := func(d time.Duration) { delays = append(delays, d) }
	halfRand := func() float64 { return 0.5 }

	calls := 0
	fn := func(attempt int) (string, error) {
		calls++
		if calls <= 2 {
			return "", &RetryableError{Message: "rate limited", StatusCode: 429}
		}
		return "ok", nil
	}

	_, err := WithRetry(fn, Options{BaseDelayMs: i64ptr(1000), Sleep: sleep, Rand: halfRand})
	if err != nil {
		t.Fatalf("WithRetry: %v", err)
	}
	// jitter = base * 0.5 * 0.5 = base * 0.25
	// attempt 0: 1000 + 250 = 1250ms; attempt 1: 2000 + 500 = 2500ms
	if delays[0] != 1250*time.Millisecond {
		t.Errorf("delays[0] = %v, want 1250ms", delays[0])
	}
	if delays[1] != 2500*time.Millisecond {
		t.Errorf("delays[1] = %v, want 2500ms", delays[1])
	}
}

func TestWithRetry_UsesRetryAfterMs(t *testing.T) {
	var delays []time.Duration
	sleep := func(d time.Duration) { delays = append(delays, d) }

	retryAfter := int64(5000)
	calls := 0
	fn := func(attempt int) (string, error) {
		calls++
		if calls == 1 {
			return "", &RetryableError{Message: "rate limited", StatusCode: 429, RetryAfterMs: &retryAfter}
		}
		return "ok", nil
	}

	_, err := WithRetry(fn, Options{BaseDelayMs: i64ptr(100), Sleep: sleep, Rand: func() float64 { return 0 }})
	if err != nil {
		t.Fatalf("WithRetry: %v", err)
	}
	if delays[0] != 5000*time.Millisecond {
		t.Errorf("delays[0] = %v, want 5000ms (retryAfterMs)", delays[0])
	}
}

func TestWithRetry_CapsDelayAtMaxDelayMs(t *testing.T) {
	var delays []time.Duration
	sleep := func(d time.Duration) { delays = append(delays, d) }
	zeroRand := func() float64 { return 0 }

	calls := 0
	fn := func(attempt int) (string, error) {
		calls++
		if calls <= 3 {
			return "", &RetryableError{Message: "rate limited", StatusCode: 429}
		}
		return "ok", nil
	}

	_, err := WithRetry(fn, Options{BaseDelayMs: i64ptr(1000), MaxDelayMs: i64ptr(3000), Sleep: sleep, Rand: zeroRand})
	if err != nil {
		t.Fatalf("WithRetry: %v", err)
	}
	// 1000, 2000, capped at 3000
	if delays[0] != 1000*time.Millisecond || delays[1] != 2000*time.Millisecond || delays[2] != 3000*time.Millisecond {
		t.Errorf("delays = %v, want [1s 2s 3s]", delays)
	}
}

func TestWithRetry_CapsRetryAfterMsAtMaxDelayMs(t *testing.T) {
	var delays []time.Duration
	sleep := func(d time.Duration) { delays = append(delays, d) }

	retryAfter := int64(120_000)
	calls := 0
	fn := func(attempt int) (string, error) {
		calls++
		if calls == 1 {
			return "", &RetryableError{Message: "rate limited", StatusCode: 429, RetryAfterMs: &retryAfter}
		}
		return "ok", nil
	}

	_, err := WithRetry(fn, Options{BaseDelayMs: i64ptr(100), MaxDelayMs: i64ptr(60_000), Sleep: sleep})
	if err != nil {
		t.Fatalf("WithRetry: %v", err)
	}
	if delays[0] != 60_000*time.Millisecond {
		t.Errorf("delays[0] = %v, want 60s (capped)", delays[0])
	}
}

func TestWithRetry_PassesAttemptNumberStartingAtZero(t *testing.T) {
	var attempts []int
	fn := func(attempt int) (string, error) {
		attempts = append(attempts, attempt)
		if attempt < 2 {
			return "", &RetryableError{Message: "retry me", StatusCode: 429}
		}
		return "done", nil
	}

	maxRetries := 3
	got, err := WithRetry(fn, Options{MaxRetries: &maxRetries, BaseDelayMs: i64ptr(0), Sleep: noSleep})
	if err != nil {
		t.Fatalf("WithRetry: %v", err)
	}
	if got != "done" {
		t.Errorf("got %q", got)
	}
	wantAttempts := []int{0, 1, 2}
	if len(attempts) != 3 || attempts[0] != 0 || attempts[1] != 1 || attempts[2] != 2 {
		t.Errorf("attempts = %v, want %v", attempts, wantAttempts)
	}
}

func TestIsRetryableStatus(t *testing.T) {
	cases := []struct {
		code int
		want bool
	}{
		{429, true}, {500, true}, {502, true}, {503, true}, {504, true},
		{200, false}, {400, false}, {401, false}, {404, false}, {501, false},
	}
	for _, c := range cases {
		if got := IsRetryableStatus(c.code); got != c.want {
			t.Errorf("IsRetryableStatus(%d) = %v, want %v", c.code, got, c.want)
		}
	}
}

func noSleep(time.Duration) {}

func i64ptr(v int64) *int64 { return &v }
