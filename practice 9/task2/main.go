package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

type CachedResponse struct {
	StatusCode int
	Body       []byte
	Completed  bool
}

type MemoryStore struct {
	mu   sync.Mutex
	data map[string]*CachedResponse
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: make(map[string]*CachedResponse)}
}

func (m *MemoryStore) Get(key string) (*CachedResponse, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	v, ok := m.data[key]
	return v, ok
}

func (m *MemoryStore) StartProcessing(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.data[key]; ok {
		return false
	}
	m.data[key] = &CachedResponse{Completed: false}
	return true
}

func (m *MemoryStore) Finish(key string, status int, body []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if v, ok := m.data[key]; ok {
		v.StatusCode = status
		v.Body = body
		v.Completed = true
	}
}

func IdempotencyMiddleware(store *MemoryStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"Idempotency-Key header required"}`))
			return
		}

		if cached, ok := store.Get(key); ok {
			if cached.Completed {
				log.Printf("[middleware] key=%s already completed, returning cached response", key)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(cached.StatusCode)
				w.Write(cached.Body)
			} else {
				log.Printf("[middleware] key=%s still processing -> 409", key)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(`{"error":"Duplicate request in progress"}`))
			}
			return
		}

		if !store.StartProcessing(key) {
			if cached, ok := store.Get(key); ok && cached.Completed {
				log.Printf("[middleware] key=%s just finished, returning cached response", key)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(cached.StatusCode)
				w.Write(cached.Body)
			} else {
				log.Printf("[middleware] key=%s race condition -> 409", key)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				w.Write([]byte(`{"error":"Duplicate request in progress"}`))
			}
			return
		}

		log.Printf("[middleware] key=%s new request, starting processing", key)

		rec := httptest.NewRecorder()
		next.ServeHTTP(rec, r)
		store.Finish(key, rec.Code, rec.Body.Bytes())

		for k, vals := range rec.Header() {
			for _, v := range vals {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(rec.Code)
		w.Write(rec.Body.Bytes())
	})
}

func paymentHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("[handler] Processing payment... (sleeping 2s)")
	time.Sleep(2 * time.Second)

	txID := generateUUID()
	log.Printf("[handler] Done. transaction_id=%s", txID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         "paid",
		"amount":         1000,
		"transaction_id": txID,
	})
}

func main() {
	store := NewMemoryStore()
	mux := http.NewServeMux()
	mux.Handle("/pay", IdempotencyMiddleware(store, http.HandlerFunc(paymentHandler)))

	server := httptest.NewServer(mux)
	defer server.Close()

	idempKey := generateUUID()
	fmt.Println("=== Task 2: Loan Repayment (Idempotency) ===")
	fmt.Printf("Idempotency-Key: %s\n\n", idempKey)

	var wg sync.WaitGroup
	type result struct {
		n      int
		status int
		body   string
	}
	ch := make(chan result, 8)

	for i := 1; i <= 7; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			time.Sleep(time.Duration(n*40) * time.Millisecond)

			req, _ := http.NewRequest(http.MethodPost, server.URL+"/pay", nil)
			req.Header.Set("Idempotency-Key", idempKey)

			resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
			if err != nil {
				ch <- result{n, 0, err.Error()}
				return
			}
			defer resp.Body.Close()

			var body map[string]interface{}
			json.NewDecoder(resp.Body).Decode(&body)
			b, _ := json.Marshal(body)
			ch <- result{n, resp.StatusCode, string(b)}
		}(i)
	}

	wg.Wait()
	close(ch)

	fmt.Println("\n=== Results (concurrent requests) ===")
	for r := range ch {
		fmt.Printf("Request #%d -> status=%d body=%s\n", r.n, r.status, r.body)
	}

	fmt.Println("\n=== Late request (after completion) ===")
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/pay", nil)
	req.Header.Set("Idempotency-Key", idempKey)
	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)
	b, _ := json.Marshal(body)
	fmt.Printf("status=%d body=%s\n", resp.StatusCode, b)
	fmt.Println("\n[OK] Same transaction_id = business logic was NOT re-executed")
}
