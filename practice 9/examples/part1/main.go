package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/url"
	"time"
)

func doSomethingUnreliable() error {
	if rand.Intn(10) < 7 {
		fmt.Println("Operation failed, retrying...")
		return errors.New("temporary failure")
	}
	fmt.Println("Operation succeeded!")
	return nil
}

func example1() {
	fmt.Println("\n--- Example 1: Naive infinite loop (anti-pattern) ---")
	var err error
	count := 0
	for {
		count++
		if count > 5 {
			fmt.Println("(stopped after 5 for demo - would loop forever)")
			break
		}
		err = doSomethingUnreliable()
		if err == nil {
			break
		}
	}
	_ = err
}

func example2() {
	fmt.Println("\n--- Example 2: Fixed delay ---")
	var err error
	const maxRetries = 5
	const delay = 200 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		err = doSomethingUnreliable()
		if err == nil {
			break
		}
		fmt.Printf("Attempt %d failed, waiting %v before next retry...\n", attempt+1, delay)
		if attempt < maxRetries-1 {
			time.Sleep(delay)
		}
	}
	if err != nil {
		fmt.Printf("Failed after %d attempts: %v\n", maxRetries, err)
	} else {
		fmt.Println("Succeeded within retry limit.")
	}
}

func example3() {
	fmt.Println("\n--- Example 3: Exponential Backoff ---")
	var err error
	const maxRetries = 5
	baseDelay := 100 * time.Millisecond
	maxDelay := 2 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		err = doSomethingUnreliable()
		if err == nil {
			break
		}
		if attempt == maxRetries-1 {
			break
		}
		backoff := baseDelay * time.Duration(math.Pow(2, float64(attempt)))
		if backoff > maxDelay {
			backoff = maxDelay
		}
		fmt.Printf("Attempt %d failed, waiting %v before next retry...\n", attempt+1, backoff)
		time.Sleep(backoff)
	}
	if err != nil {
		fmt.Printf("Failed after %d attempts\n", maxRetries)
	}
}

func example4() {
	fmt.Println("\n--- Example 4: Exponential Backoff + Full Jitter ---")
	var err error
	const maxRetries = 5
	baseDelay := 100 * time.Millisecond
	maxDelay := 2 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		err = doSomethingUnreliable()
		if err == nil {
			break
		}
		if attempt == maxRetries-1 {
			break
		}
		backoff := baseDelay * time.Duration(math.Pow(2, float64(attempt)))
		if backoff > maxDelay {
			backoff = maxDelay
		}
		jitter := time.Duration(rand.Int63n(int64(backoff) + 1))
		fmt.Printf("Attempt %d failed, waiting ~%v (backoff %v + jitter) before next retry...\n", attempt+1, jitter, backoff)
		time.Sleep(jitter)
	}
	if err != nil {
		fmt.Printf("Failed after %d attempts\n", maxRetries)
	}
}

type RetryConfig struct {
	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration
}

func Retry(ctx context.Context, cfg RetryConfig) error {
	var err error
	for attempt := 0; attempt < cfg.maxRetries; attempt++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		err = doSomethingUnreliable()
		if err == nil {
			return nil
		}
		if attempt == cfg.maxRetries-1 {
			return err
		}
		backoff := cfg.baseDelay * time.Duration(math.Pow(2, float64(attempt)))
		if backoff > cfg.maxDelay {
			backoff = cfg.maxDelay
		}
		jitter := time.Duration(rand.Int63n(int64(backoff) + 1))
		fmt.Printf("Attempt %d failed, waiting ~%v (max backoff: %v)...\n", attempt+1, jitter, backoff)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(jitter):
		}
	}
	return err
}

func example5() {
	fmt.Println("\n--- Example 5: Context + Cancellation ---")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := Retry(ctx, RetryConfig{
		maxRetries: 5,
		baseDelay:  100 * time.Millisecond,
		maxDelay:   1 * time.Second,
	})
	if err != nil {
		fmt.Printf("Failed: %v\n", err)
	} else {
		fmt.Println("Success!")
	}
}

type timeoutErr struct{}

func (t *timeoutErr) Error() string   { return "i/o timeout" }
func (t *timeoutErr) Timeout() bool   { return true }
func (t *timeoutErr) Temporary() bool { return true }

func isRetryable(err error) bool {
	var urlErr *url.Error
	if errors.As(err, &urlErr) && urlErr.Timeout() {
		return true
	}
	return false
}

func example6() {
	fmt.Println("\n--- Example 6: Error Filtering ---")
	permanent := errors.New("404 not found")
	temporary := &url.Error{Op: "Get", URL: "http://example.com", Err: &timeoutErr{}}

	fmt.Printf("permanent error -> isRetryable: %v\n", isRetryable(permanent))
	fmt.Printf("timeout error   -> isRetryable: %v\n", isRetryable(temporary))
}

func main() {
	rand.Seed(time.Now().UnixNano())
	fmt.Println("=== Part 1: Retry Patterns ===")
	example1()
	example2()
	example3()
	example4()
	example5()
	example6()
}
