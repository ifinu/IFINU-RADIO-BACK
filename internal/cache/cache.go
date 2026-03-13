package cache

import (
	"sync"
	"time"

	"github.com/ifinu/radio-api/internal/models"
)

type CacheItem struct {
	Value      *models.Radio
	Expiration time.Time
}

type RadioCache struct {
	items sync.Map
	ttl   time.Duration
}

func NewRadioCache(ttl time.Duration) *RadioCache {
	cache := &RadioCache{
		ttl: ttl,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

func (c *RadioCache) Set(id uint, radio *models.Radio) {
	c.items.Store(id, &CacheItem{
		Value:      radio,
		Expiration: time.Now().Add(c.ttl),
	})
}

func (c *RadioCache) Get(id uint) (*models.Radio, bool) {
	value, ok := c.items.Load(id)
	if !ok {
		return nil, false
	}

	item := value.(*CacheItem)
	if time.Now().After(item.Expiration) {
		c.items.Delete(id)
		return nil, false
	}

	return item.Value, true
}

func (c *RadioCache) Delete(id uint) {
	c.items.Delete(id)
}

func (c *RadioCache) Clear() {
	c.items.Range(func(key, value interface{}) bool {
		c.items.Delete(key)
		return true
	})
}

func (c *RadioCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		c.items.Range(func(key, value interface{}) bool {
			item := value.(*CacheItem)
			if now.After(item.Expiration) {
				c.items.Delete(key)
			}
			return true
		})
	}
}
