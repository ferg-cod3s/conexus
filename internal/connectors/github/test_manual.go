package github

import (
	"fmt"
	"net/http"
	"time"
)

func testUpdateRateLimit() {
	client := NewHTTPClient("test-token")
	
	fmt.Printf("=== Manual Test ===\n")
	fmt.Printf("Initial remaining: %d\n", client.rateLimiter.Remaining())
	
	resp := &http.Response{
		Header: http.Header{
			"X-RateLimit-Remaining": []string{"45"},
			"X-RateLimit-Reset":     []string{fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix())},
		},
	}
	
	fmt.Printf("X-RateLimit-Remaining header: '%s'\n", resp.Header.Get("X-RateLimit-Remaining"))
	
	client.updateRateLimit(resp)
	
	fmt.Printf("After updateRateLimit: %d (expected 45)\n", client.rateLimiter.Remaining())
}
