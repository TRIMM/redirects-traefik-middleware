package redirects_traefik_middleware

import (
	"sync"
	"time"
)

type Cache struct {
	data        map[string]interface{}
	expiration  map[string]time.Time
	mutex       sync.RWMutex
	defaultTTL  time.Duration
	cleanupTick time.Duration
}

func NewCache(defaultTTL time.Duration, cleanupTick time.Duration) *Cache {
	cache := &Cache{
		data:        make(map[string]interface{}),
		expiration:  make(map[string]time.Time),
		defaultTTL:  defaultTTL,
		cleanupTick: cleanupTick,
	}
	go cache.startCleanup()

	return cache
}

func (c *Cache) startCleanup() {
	ticker := time.NewTicker(c.cleanupTick)
	for {
		select {
		case <-ticker.C:
			c.cleanup()
		}
	}
}

func (c *Cache) cleanup() {
	currentTime := time.Now()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	for key, expirationTime := range c.expiration {
		if currentTime.After(expirationTime) {
			delete(c.data, key)
			delete(c.expiration, key)
		}
	}
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = value
	c.expiration[key] = time.Now().Add(ttl)
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	value, ok := c.data[key]
	return value, ok
}
