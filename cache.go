package prehit

import (
	"sync"
	"time"

	"go.melnyk.org/mlog"
	"go.melnyk.org/mlog/nolog"
)

// cacheItem is a single item in the cache.
type cacheItem[K comparable, V any] struct {
	prev       *cacheItem[K, V]
	next       *cacheItem[K, V]
	key        K
	value      V
	expiration time.Time
}

// Cache is a simple in-memory cache.
type Cache[K comparable, V any] struct {
	logger  mlog.Logger
	index   map[K]*cacheItem[K, V]
	head    *cacheItem[K, V]
	tail    *cacheItem[K, V]
	maxsize uint
	size    uint
	metrics Metrics
	mutex   sync.RWMutex
	pool    *sync.Pool
}

// NewCache creates a new cache.
func NewCache[K comparable, V any](o ...Option) *Cache[K, V] {
	local := &options{
		logger:  nolog.NewLogbook().Joiner().Join(""), // default logger
		maxsize: 1000,                                 // default max size
		metrics: &nometrics{},                         // no metrics by default
	}

	for _, option := range o {
		option.apply(local)
	}

	return &Cache[K, V]{
		logger:  local.logger,
		index:   make(map[K]*cacheItem[K, V], local.maxsize),
		maxsize: local.maxsize,
		metrics: local.metrics,
		size:    0,
		head:    nil,
		tail:    nil,
		pool: &sync.Pool{
			New: func() interface{} {
				return new(cacheItem[K, V])
			},
		},
	}
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	now := time.Now()
	c.mutex.RLock()

	if item, found := c.index[key]; found {
		if item != nil {
			if item.expiration.After(now) {
				value := item.value
				movetohead := (item.next == nil) && (c.size > 1)
				c.mutex.RUnlock()

				if movetohead { // last Item and more than one item
					c.mutex.Lock()
					c.movetohead(key)
					c.mutex.Unlock()
				}
				c.metrics.Hit()
				return value, true
			} else {
				// remove expired element - mutex relock is needed
				c.mutex.RUnlock()
				c.mutex.Lock()
				c.deleteexpired(key, now)
				c.mutex.Unlock()
				c.metrics.Miss()
				c.metrics.Evict()
				c.metrics.Delete()
				return *new(V), false
			}
		} else {
			c.logger.Warning("Inconsistency in the cache structure - cache item cannot be nil")
			c.metrics.Error()
		}
	}

	c.mutex.RUnlock()
	c.metrics.Miss()

	return *new(V), false
}

func (c *Cache[K, V]) movetohead(key K) {
	if item, found := c.index[key]; found { // it can be changes in cache, extra check is needed
		if item != nil {
			if item.next == nil { // move item to the head
				c.tail = item.prev
				if c.tail != nil {
					c.tail.next = nil
				}
				item.prev = nil
				item.next = c.head
				if c.head != nil {
					c.head.prev = item
				}
				c.head = item
			}
		}
	}
}

func (c *Cache[K, V]) deleteexpired(key K, tm time.Time) {
	if item, found := c.index[key]; found { // it can be changes in cache, extra check is needed
		if item != nil {
			if item.expiration.Before(tm) {
				if item.next != nil {
					item.next.prev = item.prev
				}
				if item.prev != nil {
					item.prev.next = item.next
				}
				if c.tail == item {
					c.tail = item.prev
				}
				if c.head == item {
					c.head = item.next
				}
				item.prev = nil
				item.next = nil
			}
		} else {
			c.logger.Warning("Inconsistency in the cache structure - item element cannot be nil")
			c.metrics.Error()
		}
		delete(c.index, key)
		c.pool.Put(item)
		if c.size > 0 {
			c.size--
		} else {
			c.logger.Warning("Inconsistency in the cache structure - more items deleted than expected")
			c.metrics.Error()
		}
	}
}

// Set stores a value for a key.
func (c *Cache[K, V]) Set(key K, v V, ttl time.Duration) {
	c.logger.Verbose("Set cache")
	expiration := time.Now().Add(ttl)
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if item, found := c.index[key]; found {
		item.value = v
		item.expiration = expiration
		if item.next == nil { // last item
			if c.size > 1 { // more than one item
				// move item from the tail to the head
				c.tail = item.prev
				c.tail.next = nil
				item.prev = nil
				item.next = c.head
				c.head.prev = item
				c.head = item
			}
		}
		c.metrics.Update()
		return
	}

	if c.size == c.maxsize { // we reached max size,
		// remove the last item
		if c.tail != nil {
			item := c.tail
			c.tail = c.tail.prev
			if c.tail != nil { // more than one item
				c.tail.next = nil
			} else { // list is empty
				c.head = nil
			}
			c.size--
			delete(c.index, item.key)
			item.prev = nil
			c.pool.Put(item)
			c.metrics.Delete()
			c.metrics.Evict()
		}
	}

	// add new item to the head
	item := c.pool.Get().(*cacheItem[K, V])
	item.key = key
	item.value = v
	item.expiration = expiration
	item.prev = nil
	item.next = c.head

	if c.head != nil { // list is not empty
		c.head.prev = item
	} else { // list is empty and we need to set the tail also
		c.tail = item
	}

	c.head = item
	c.index[key] = item
	c.size++
	c.metrics.Add()
}

// Delete removes a key from the cache.
func (c *Cache[K, V]) Delete(key ...K) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for _, k := range key {
		if item, found := c.index[k]; found {
			if item != nil {
				if item.next != nil {
					item.next.prev = item.prev
				}
				if item.prev != nil {
					item.prev.next = item.next
				}
				if c.tail == item {
					c.tail = item.prev
				}
				if c.head == item {
					c.head = item.next
				}
				item.prev = nil
				item.next = nil

				c.pool.Put(item)
			}
			delete(c.index, k)
			c.metrics.Delete()
			if c.size > 0 {
				c.size--
			} else {
				c.logger.Warning("Inconsistency in the cache structure - more items deleted than expected")
				c.metrics.Error()
			}
		}
	}
}

// Reset clears the cache.
func (c *Cache[K, V]) Reset() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for c.head != nil {
		item := c.head
		c.head = c.head.next
		item.prev = nil
		item.next = nil
		c.pool.Put(item)
		c.metrics.Delete()
	}

	// recreate index
	c.index = make(map[K]*cacheItem[K, V], c.maxsize)
	c.tail = nil
	c.size = 0

	return nil
}
