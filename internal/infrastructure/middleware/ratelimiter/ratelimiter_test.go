// internal/infrastructure/middleware/ratelimiter/ratelimiter_test.go
package ratelimiter

import (
	"Auth/pkg/logger"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	// Create a real logger instance for testing
	logInstance, err := logger.NewSimpleLogger("info")
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer logInstance.Close()

	// Create a new rate limiter: 3 requests per second
	limiter := NewRateLimiter(3, time.Second, logInstance)

	// Create a simple test handler that returns 200 OK
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Apply the rate limiter middleware
	handler := limiter.Middleware(testHandler)

	// Create a test server
	server := httptest.NewServer(handler)
	defer server.Close()

	// Make multiple requests from the same IP
	client := &http.Client{}

	// First 3 requests should succeed
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("GET", server.URL, nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.1") // Set a consistent IP

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status OK, got %v for request %d", resp.StatusCode, i+1)
		}
		resp.Body.Close()
	}

	// Fourth request should be blocked
	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Expected status 429 Too Many Requests, got %v", resp.StatusCode)
	}

	// Test different IP - should succeed
	req, _ = http.NewRequest("GET", server.URL, nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.2") // Different IP

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK for different IP, got %v", resp.StatusCode)
	}

	// Wait for the time window to expire
	time.Sleep(time.Second)

	// After waiting, should be able to make another request
	req, _ = http.NewRequest("GET", server.URL, nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1") // Original IP

	resp, err = client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK after waiting, got %v", resp.StatusCode)
	}
}
