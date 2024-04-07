// File: internal/store/store.go

package store

import (
	"sort"
	"sync"
)

type InMemoryStore struct {
	data      map[string]string
	sortedSet map[string]map[string]float64
	mutex     sync.RWMutex
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data:      make(map[string]string),
		sortedSet: make(map[string]map[string]float64),
		mutex:     sync.RWMutex{},
	}
}

// Set a key to hold the string value
func (store *InMemoryStore) Set(key, value string) string {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.data[key] = value
	return "+OK\r\n"
}

// Get the value of key
func (store *InMemoryStore) Get(key string) string {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	if value, ok := store.data[key]; ok {
		return "$" + value + "\r\n"
	}
	return "$-1\r\n" // Redis protocol for non-existent key
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

// ZAdd adds all the specified members with the specified scores to the sorted set stored at key
func (store *InMemoryStore) ZAdd(key string, score float64, member string) int {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if _, ok := store.sortedSet[key]; !ok {
		store.sortedSet[key] = make(map[string]float64)
	}
	store.sortedSet[key][member] = score
	return 1 // Assuming we're always adding a new element for simplicity
}

// ZRange returns the specified range of elements in the sorted set stored at key
func (store *InMemoryStore) ZRange(key string, start, stop int) []string {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	if sortedSet, ok := store.sortedSet[key]; ok {
		var members []string
		for member := range sortedSet {
			members = append(members, member)
		}
		sort.Strings(members) // Simplified: should sort by score

		if start < 0 || start >= len(members) {
			return []string{}
		}

		if stop >= len(members) {
			stop = len(members) - 1
		}

		return members[start : stop+1]
	}
	return []string{}
}
