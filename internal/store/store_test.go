package store

import (
	"fmt"
	"testing"
	"time"
)

func TestSetAndGet(t *testing.T) {
	store := NewInMemoryStore()
	key, value := "testkey", "testvalue"

	// Test setting a value without flags
	result := store.Set(key, value)
	if result != "+1\r\n" {
		t.Errorf("Set(%q, %q) = %q, want %q", key, value, result, "+1\r\n")
	}

	// Test getting the set value
	gotValue := store.Get(key)
	if gotValue != fmt.Sprintf("$%d\r\n%s\r\n\r\n", len(value), value) {
		t.Errorf("Get(%q) = %q, want %q", key, gotValue, value)
	}

	// Test setting a value with NX flag (should not set as key already exists)
	result = store.Set(key, "newvalue", "NX")
	if result != "+0\r\n" {
		t.Errorf("Set(%q, %q, NX) = %q, want %q", key, "newvalue", result, "+0\r\n")
	}

	// Test setting a value with XX flag (should set as key already exists)
	result = store.Set(key, "newvalue", "XX")
	if result != "+1\r\n" {
		t.Errorf("Set(%q, %q, XX) = %q, want %q", key, "newvalue", result, "+1\r\n")
	}

	// Test non-existent key
	nonExistentKey := "nonexistent"
	if got := store.Get(nonExistentKey); got != "$-1\r\n\r\n" {
		t.Errorf("Get(%q) = %q, want %q", nonExistentKey, got, "$-1\r\n\r\n")
	}

	// Test setting a new key with EX (expiration) flag
	newKey, newValue := "tempkey", "tempvalue"
	result = store.Set(newKey, newValue, "EX10") // 10 seconds expiration
	if result != "+1\r\n" {
		t.Errorf("Set(%q, %q, EX10) = %q, want %q", newKey, newValue, result, "+1\r\n")
	}
}

func TestDel(t *testing.T) {
	store := NewInMemoryStore()
	key := "testkey"

	store.Set(key, "value")
	if deleted := store.Del([]string{key}); deleted != 1 {
		t.Errorf("Del(%q) = %d, want %d", key, deleted, 1)
	}

	// Test deletion of a non-existent key
	if deleted := store.Del([]string{key}); deleted != 0 {
		t.Errorf("Del(%q) = %d, want %d", key, deleted, 0)
	}
}

func TestKeys(t *testing.T) {
	store := NewInMemoryStore()
	keyPattern := "user:*"
	store.Set("user:1", "Alice")
	store.Set("user:2", "Bob")
	store.Set("order:1", "Order1")

	keys := store.Keys(keyPattern)
	if len(keys) != 2 {
		t.Errorf("Keys(%q) returned %d keys, want %d", keyPattern, len(keys), 2)
	}
}

func TestExpireAndTTL(t *testing.T) {
	store := NewInMemoryStore()
	key := "tempkey"

	store.Set(key, "value")
	store.Expire(key, 1) // 1 second expiration

	time.Sleep(2 * time.Second) // Wait for the key to expire

	ttl := store.TTL(key)
	if ttl != -2 {
		t.Errorf("TTL(%q) = %d, want %d", key, ttl, -2)
	}
}

func TestZAddAndZRange(t *testing.T) {
	store := NewInMemoryStore()
	key := "sortedset"
	store.ZAdd(key, 1, "one")
	store.ZAdd(key, 2, "two")
	store.ZAdd(key, 3, "three")

	members := store.ZRange(key, 0, -1) // Get all members
	if len(members) != 3 {
		t.Errorf("ZRange(%q, 0, -1) returned %d members, want %d", key, len(members), 3)
	}
}
