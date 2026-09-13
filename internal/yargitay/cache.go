// SPDX-License-Identifier: AGPL-3.0-only
package yargitay

import (
	"container/list"
	"sync"
	"time"
)

type cacheEntry struct {
	key     string
	data    []byte
	expires time.Time
}
type memoryCache struct {
	mu                        sync.Mutex
	items                     map[string]*list.Element
	lru                       *list.List
	bytes, maxBytes, maxItems int
}

func newCache(items, bytes int) *memoryCache {
	return &memoryCache{items: map[string]*list.Element{}, lru: list.New(), maxBytes: bytes, maxItems: items}
}
func (c *memoryCache) remove(e *list.Element) {
	v := e.Value.(cacheEntry)
	delete(c.items, v.key)
	c.bytes -= len(v.data)
	c.lru.Remove(e)
}
func (c *memoryCache) get(key string) []byte {
	c.mu.Lock()
	defer c.mu.Unlock()
	e := c.items[key]
	if e == nil {
		return nil
	}
	v := e.Value.(cacheEntry)
	if !time.Now().Before(v.expires) {
		c.remove(e)
		return nil
	}
	c.lru.MoveToFront(e)
	return append([]byte(nil), v.data...)
}
func (c *memoryCache) set(key string, data []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if e := c.items[key]; e != nil {
		c.remove(e)
	}
	if ttl <= 0 || len(data) > c.maxBytes {
		return
	}
	for _, e := range c.items {
		if !time.Now().Before(e.Value.(cacheEntry).expires) {
			c.remove(e)
		}
	}
	for c.lru.Len() > 0 && (c.lru.Len() >= c.maxItems || c.bytes+len(data) > c.maxBytes) {
		c.remove(c.lru.Back())
	}
	c.items[key] = c.lru.PushFront(cacheEntry{key, append([]byte(nil), data...), time.Now().Add(ttl)})
	c.bytes += len(data)
}
