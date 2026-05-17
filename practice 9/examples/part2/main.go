package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
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

func example1() {
	fmt.Println("\n--- Example 1: Problem without idempotency (double charge) ---")
	balance := 10000

	debit := func(amount int) {
		balance -= amount
		fmt.Printf("Debited $%d -> balance: $%d\n", amount, balance)
	}

	fmt.Printf("Initial balance: $%d\n", balance)
	fmt.Println("Request 1 (real):")
	debit(1000)
	fmt.Println("Request 2 (retry, no idempotency):")
	debit(1000)
	fmt.Printf("Final balance: $%d (WRONG - double charged!)\n", balance)
}

func example2() {
	fmt.Println("\n--- Example 2: With idempotency key ---")
	balance := 10000
	processed := make(map[string]bool)

	debit := func(key string, amount int) {
		if processed[key] {
			fmt.Printf("Key already seen, skipping. Balance: $%d\n", balance)
			return
		}
		processed[key] = true
		balance -= amount
		fmt.Printf("Debited $%d -> balance: $%d\n", amount, balance)
	}

	key := generateUUID()
	fmt.Printf("Initial balance: $%d\n", balance)
	fmt.Printf("Request 1 (key=%s...):\n", key[:8])
	debit(key, 1000)
	fmt.Println("Request 2 (same key - retry):")
	debit(key, 1000)
	fmt.Printf("Final balance: $%d (CORRECT)\n", balance)
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

func idempotencyMiddleware(store *MemoryStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			http.Error(w, "Idempotency-Key header required", http.StatusBadRequest)
			return
		}
		if cached, ok := store.Get(key); ok {
			if cached.Completed {
				w.WriteHeader(cached.StatusCode)
				w.Write(cached.Body)
			} else {
				http.Error(w, "Duplicate request in progress", http.StatusConflict)
			}
			return
		}
		if !store.StartProcessing(key) {
			if cached, ok := store.Get(key); ok && cached.Completed {
				w.WriteHeader(cached.StatusCode)
				w.Write(cached.Body)
			} else {
				http.Error(w, "Duplicate request in progress", http.StatusConflict)
			}
			return
		}
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

func example3() {
	fmt.Println("\n--- Example 3: Middleware demo (3 requests, same key) ---")

	store := NewMemoryStore()
	calls := 0

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"id": generateUUID()[:8],
		})
	})

	server := httptest.NewServer(idempotencyMiddleware(store, handler))
	defer server.Close()

	key := generateUUID()
	for i := 1; i <= 3; i++ {
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/resource", nil)
		req.Header.Set("Idempotency-Key", key)
		resp, _ := http.DefaultClient.Do(req)
		var body map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		fmt.Printf("Request #%d -> status=%d body=%v\n", i, resp.StatusCode, body)
	}
	fmt.Printf("Handler called %d time(s) (expected: 1)\n", calls)
}

func example4() {
	fmt.Println("\n--- Example 4: Redis SETNX pattern simulation ---")

	type store struct {
		mu   sync.Mutex
		data map[string]string
	}
	s := &store{data: make(map[string]string)}

	setNX := func(key, val string) bool {
		s.mu.Lock()
		defer s.mu.Unlock()
		if _, ok := s.data[key]; ok {
			return false
		}
		s.data[key] = val
		return true
	}
	get := func(key string) string {
		s.mu.Lock()
		defer s.mu.Unlock()
		return s.data[key]
	}
	set := func(key, val string) {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.data[key] = val
	}

	key := "idempotency:pay-abc123"

	if setNX(key, "processing") {
		fmt.Println("Request 1: acquired, processing...")
		time.Sleep(50 * time.Millisecond)
		set(key, `{"status":"paid","amount":1000}`)
		fmt.Println("Request 1: saved result to store")
	}

	if !setNX(key, "processing") {
		val := get(key)
		if val == "processing" {
			fmt.Println("Request 2: still processing -> 409 Conflict")
		} else {
			fmt.Printf("Request 2: found result -> %s\n", val)
		}
	}
}

func main() {
	fmt.Println("=== Part 2: Idempotency Examples ===")
	example1()
	example2()
	example3()
	example4()
}
