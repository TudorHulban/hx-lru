package hxlru

import (
	"container/list"
	"fmt"
	"strings"
	"sync"
	"time"

	goerrors "github.com/TudorHulban/go-errors"
)

type itemOne[V any] struct {
	keyPtr *list.Element // holds key of item in cache.

	payload      V
	timestampTTL int64
}

type CacheOneLRU[K comparable, V any] struct {
	Queue *list.List
	Cache map[K]*itemOne[V]

	mu sync.Mutex

	ttl      time.Duration
	capacity uint16
}

func NewCacheOneLRU[K comparable, V any](params *ParamsNewCacheLRU) *CacheOneLRU[K, V] {
	return &CacheOneLRU[K, V]{
		Queue: list.New(),
		Cache: make(map[K]*itemOne[V]),

		capacity: params.Capacity,
		ttl:      params.TTL,
	}
}

func (c *CacheOneLRU[K, V]) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Capacity: %d\n", c.capacity))
	sb.WriteString("Cached:\n")

	for key, item := range c.Cache {
		sb.WriteString(
			fmt.Sprintf(
				"key: %v, value: %v\n",

				key,
				item.payload,
			),
		)
	}

	return sb.String()
}

func (c *CacheOneLRU[K, V]) Put(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if node, exists := c.Cache[key]; !exists {
		if int(c.capacity) == len(c.Cache) {
			evicted := c.Queue.Back()
			c.Queue.Remove(evicted)

			delete(
				c.Cache,
				evicted.Value.(K),
			)
		}

		c.Cache[key] = &itemOne[V]{
			keyPtr:  c.Queue.PushFront(key),
			payload: value,
		}

		return
	} else {
		node.payload = value

		c.Cache[key] = node
		c.Queue.MoveToFront(node.keyPtr)
	}
}

func (c *CacheOneLRU[K, V]) PutTTL(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, exists := c.Cache[key]
	if !exists {
		if int(c.capacity) == len(c.Cache) {
			evicted := c.Queue.Back()
			c.Queue.Remove(evicted)

			delete(
				c.Cache,
				evicted.Value.(K),
			)
		}

		c.Cache[key] = &itemOne[V]{
			keyPtr:  c.Queue.PushFront(key),
			payload: value,

			timestampTTL: time.Now().
				Add(c.ttl).
				UnixNano(),
		}

		return
	}

	node.payload = value
	node.timestampTTL = time.Now().
		Add(c.ttl).
		UnixNano()

	c.Cache[key] = node

	c.Queue.MoveToFront(node.keyPtr)
}

func (c *CacheOneLRU[K, V]) Get(key K) (*V, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, exists := c.Cache[key]; exists {
		if item.timestampTTL > 0 && time.Now().UnixNano() >= item.timestampTTL {
			go func() {
				_ = c.Delete(key)
			}()

			return nil,
				goerrors.ErrEntryNotFound{
					Key: key,
				}
		}

		c.Queue.MoveToFront(item.keyPtr)

		return &item.payload,
			nil
	}

	return nil,
		goerrors.ErrEntryNotFound{
			Key: key,
		}
}

func (c *CacheOneLRU[K, V]) Delete(key K) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, exists := c.Cache[key]; exists {
		c.Queue.Remove(item.keyPtr)

		delete(c.Cache, key)

		return nil
	}

	return goerrors.ErrEntryNotFound{
		Key: key,
	}
}

func (c *CacheOneLRU[K, V]) DeleteSilent(key K) {
	c.mu.Lock()

	if item, exists := c.Cache[key]; exists {
		c.Queue.Remove(item.keyPtr)

		delete(c.Cache, key)
	}

	c.mu.Unlock()
}
