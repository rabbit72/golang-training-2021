// Implement in-memory cache. It must be safe for concurrent usage.
// TTL means duration while key is valid. Invalidation should happen automatically.
// After reading a key its TTL should be increased up to current time + TTL.

package cache

import (
	"sync"
	"time"
)

type Cache interface {
	Set(key string, value interface{}, ttl time.Duration)
	Get(key string) (interface{}, bool)
	Delete(key string)
}

type valueContainter struct {
	value          interface{}
	expirationTime time.Time
	ttl            time.Duration
}

func (container *valueContainter) isExpired() bool {
	return time.Now().After(container.expirationTime)
}

func (container *valueContainter) updateExpirationTime(ttl time.Duration) {
	container.expirationTime = time.Now().Add(ttl * time.Nanosecond)
}

type SimpleCache struct {
	storage map[string]valueContainter
	mu      sync.RWMutex
	done    chan struct{}
}

// constructor for SimpleCache
func NewSimpleCache(invalidationPeriod time.Duration) *SimpleCache {
	cache := &SimpleCache{
		storage: make(map[string]valueContainter),
		done:    make(chan struct{}),
	}
	go runInvalidation(cache, invalidationPeriod)
	return cache
}

func (cache *SimpleCache) Get(key string) (interface{}, bool) {
	cache.mu.RLock()
	container, ok := cache.storage[key]
	cache.mu.RUnlock()
	
	if !ok || container.isExpired(){
		return nil, false
	}

	container.updateExpirationTime(container.ttl)
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.storage[key] = container
	return container.value, true
}

func (cache *SimpleCache) Set(key string, value interface{}, ttl time.Duration) {
	cache.mu.RLock()
	tmpContainer := cache.storage[key]
	cache.mu.RUnlock()

	tmpContainer.value = value
	tmpContainer.ttl = ttl
	tmpContainer.updateExpirationTime(ttl)
	
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.storage[key] = tmpContainer
}

func (cache *SimpleCache) Delete(key string) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	delete(cache.storage, key)
}

func (cache *SimpleCache) Stop() {
	close(cache.done)
}

func (cache *SimpleCache) invalidateExpired() {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	for key, container := range cache.storage {
		if container.isExpired() {
			delete(cache.storage, key)
		}
	}
}

func runInvalidation(cache *SimpleCache, invalidationPeriod time.Duration) {
	ticker := time.NewTicker(invalidationPeriod)

	for {
		select {
		case <-ticker.C:
			cache.invalidateExpired()
		case <-cache.done:
			return
		}
	}

}
