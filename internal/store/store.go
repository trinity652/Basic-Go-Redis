// File: internal/store/store.go

package store

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type InMemoryStore struct {
	data       map[string]string
	sortedSet  map[string]map[string]float64
	expiration map[string]time.Time
	mutex      sync.RWMutex
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data:       make(map[string]string),
		sortedSet:  make(map[string]map[string]float64),
		expiration: make(map[string]time.Time),
		mutex:      sync.RWMutex{},
	}
}

func matchPattern(key, pattern string) bool {
	// Convert the Redis pattern to a regular expression pattern
	var regexPatternBuilder strings.Builder
	for i := 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '*':
			regexPatternBuilder.WriteString(".*")
		case '?':
			regexPatternBuilder.WriteString(".")
		case '[', ']', '(', ')', '{', '}', '^', '$', '.', '|', '+', '\\':
			// Escape regex special characters
			regexPatternBuilder.WriteString("\\")
			regexPatternBuilder.WriteByte(pattern[i])
		default:
			regexPatternBuilder.WriteByte(pattern[i])
		}
	}

	regexPattern, err := regexp.Compile("^" + regexPatternBuilder.String() + "$")
	if err != nil {
		// In case of regex compilation error, fallback to simple comparison
		return key == pattern
	}

	return regexPattern.MatchString(key)
}

// Set a key to hold the string value
func (store *InMemoryStore) Set(key, value string, flags ...string) string {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	// Initialize variables for flags
	setIfExists := false
	setIfNotExists := false
	var expiration int

	// Parse flags
	for _, flag := range flags {
		switch {
		case flag == "NX":
			setIfNotExists = true
		case flag == "XX":
			setIfExists = true
		case strings.HasPrefix(flag, "EX"):
			if ex, err := strconv.Atoi(flag[2:]); err == nil {
				expiration = ex
			}
		}
	}

	// Check conditions for NX and XX
	if setIfNotExists && store.exists(key) {
		return "+0\r\n" // Key exists, do not set
	}
	if setIfExists && !store.exists(key) {
		return "+0\r\n" // Key does not exist, do not set
	}

	// Set the key
	store.data[key] = value

	// Set expiration if needed
	if expiration > 0 {
		expirationTime := time.Now().Add(time.Duration(expiration) * time.Second)
		store.expiration[key] = expirationTime
	}

	return "+1\r\n" // Success
}

// Helper function to check if a key exists
func (store *InMemoryStore) exists(key string) bool {
	_, exists := store.data[key]
	return exists
}

// Get the value of key
func (store *InMemoryStore) Get(key string) string {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	if value, ok := store.data[key]; ok {
		return fmt.Sprintf("$%d\r\n%s\r\n\r\n", len(value), value)
	}
	return "$-1\r\n\r\n" // Correct RESP format for non-existent key
}

// Del removes the specified keys
func (store *InMemoryStore) Del(keys []string) int {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	count := 0
	for _, key := range keys {
		if _, exists := store.data[key]; exists {
			delete(store.data, key)
			count++
		}
	}
	return count
}

func (store *InMemoryStore) Keys(pattern string) []string {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	var keys []string
	for key := range store.data {
		if matchPattern(key, pattern) {
			keys = append(keys, key)
		}
	}
	return keys
}

func (store *InMemoryStore) Expire(key string, seconds int) int {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, exists := store.data[key]; exists {
		expirationTime := time.Now().Add(time.Duration(seconds) * time.Second)
		store.expiration[key] = expirationTime
		return 1
	}
	return 0
}

func (store *InMemoryStore) TTL(key string) int {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	if expirationTime, exists := store.expiration[key]; exists {
		if time.Now().After(expirationTime) {
			return -2 // Key has expired
		}
		return int(time.Until(expirationTime).Seconds())
	}
	return -1 // Key does not exist or has no associated expiration
}

// ZAdd adds all the specified members with the specified scores to the sorted set stored at key
func (store *InMemoryStore) ZAdd(key string, score float64, member string) int {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, ok := store.sortedSet[key]; !ok {
		store.sortedSet[key] = make(map[string]float64)
	}

	_, memberExists := store.sortedSet[key][member]
	store.sortedSet[key][member] = score

	if memberExists {
		return 0
	}
	return 1
}

// ZRange returns the specified range of elements in the sorted set stored at key
func (store *InMemoryStore) ZRange(key string, start, stop int) []string {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	if sortedSet, ok := store.sortedSet[key]; ok {
		// Convert map to slice of key-value pairs and sort by score
		members := make([]string, 0, len(sortedSet))
		for member, score := range sortedSet {
			members = append(members, fmt.Sprintf("%f:%s", score, member))
		}
		sort.Strings(members)

		// Extract just the member names
		for i, member := range members {
			parts := strings.SplitN(member, ":", 2)
			members[i] = parts[1]
		}

		// Handle negative indices and range exceeding size
		if start < 0 {
			start = len(members) + start
		}
		if start < 0 {
			start = 0
		}
		if stop < 0 {
			stop = len(members) + stop
		}
		if stop >= len(members) {
			stop = len(members) - 1
		}

		if start > stop {
			return []string{}
		}

		return members[start : stop+1]
	}
	return []string{}
}
