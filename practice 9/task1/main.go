package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"
)

type RetryConfig struct {
	MaxRetries int
	BaseDelay  time.Duration
	MaxDelay   time.Duration
}

type PaymentClient struct {
	cfg        RetryConfig
	httpClient *http.Client
}

func NewPaymentClient(cfg RetryConfig) *PaymentClient {
	return &PaymentClient{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func IsRetryable(resp *http.Response, err error) bool {
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) && urlErr.Timeout() {
			return true
		}
		return true
	}
	switch resp.StatusCode {
	case 429, 500, 502, 503, 504:
		return true
	}
	return false
}

func CalculateBackoff(attempt int, base, max time.Duration) time.Duration {
	backoff := base * time.Duration(math.Pow(2, float64(attempt)))
	if backoff > max {
		backoff = max
	}
	return time.Duration(rand.Int63n(int64(backoff) + 1))
}

func (c *PaymentClient) ExecutePayment(ctx context.Context, serverURL string) error {
	var lastErr error

	for attempt := 0; attempt < c.cfg.MaxRetries; attempt++ {
		if ctx.Err() != nil {
			return fmt.Errorf("context cancelled: %w", ctx.Err())
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, serverURL+"/pay", nil)
		if err != nil {
			return err
		}

		resp, err := c.httpClient.Do(req)

		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			fmt.Printf("Attempt %d: Success!\n", attempt+1)
			return nil
		}

		if !IsRetryable(resp, err) {
			if resp != nil {
				resp.Body.Close()
				return fmt.Errorf("non-retryable error: status %d", resp.StatusCode)
			}
			return fmt.Errorf("non-retryable: %w", err)
		}

		if resp != nil {
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
			resp.Body.Close()
		} else {
			lastErr = err
		}

		if attempt == c.cfg.MaxRetries-1 {
			break
		}

		wait := CalculateBackoff(attempt, c.cfg.BaseDelay, c.cfg.MaxDelay)
		fmt.Printf("Attempt %d failed: waiting %v...\n", attempt+1, wait.Round(time.Millisecond))

		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled while waiting: %w", ctx.Err())
		case <-time.After(wait):
		}
	}

	return fmt.Errorf("all %d attempts failed: %w", c.cfg.MaxRetries, lastErr)
}

func main() {
	requestCount := 0

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		if requestCount <= 3 {
			fmt.Printf("[Server] Request #%d -> 503 Service Unavailable\n", requestCount)
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		fmt.Printf("[Server] Request #%d -> 200 OK\n", requestCount)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}))
	defer server.Close()

	client := NewPaymentClient(RetryConfig{
		MaxRetries: 5,
		BaseDelay:  500 * time.Millisecond,
		MaxDelay:   5 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println("=== Task 1: Resilient HTTP Client ===")
	err := client.ExecutePayment(ctx, server.URL)
	if err != nil {
		fmt.Printf("Payment failed: %v\n", err)
	} else {
		fmt.Println("Payment completed successfully!")
	}
}
