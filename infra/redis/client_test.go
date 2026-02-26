package redis

import (
	"testing"
)

func TestNewRedisClient_InvalidAddr(t *testing.T) {
	_, err := NewRedisClient("localhost:99999")
	if err == nil {
		t.Fatal("expected error for invalid redis address, got nil")
	}
}
