package storage

import (
	"fmt"
	"github.com/dfds/confluent-gateway/models"
)

func NewClusterCache(items []*models.Cluster) *Cache[models.ClusterId, *models.Cluster] {
	return NewCache(items, func(cluster *models.Cluster) models.ClusterId {
		return cluster.ClusterId
	})
}

type Cache[K ~string, T any] struct {
	cache map[K]T
}

func NewCache[K ~string, T any](items []T, key func(T) K) *Cache[K, T] {
	cache := map[K]T{}

	for _, item := range items {
		cache[key(item)] = item
	}

	return &Cache[K, T]{cache: cache}
}

type ErrKeyNotFound string

func (s ErrKeyNotFound) Error() string {
	return fmt.Sprintf("key %q not found", string(s))
}

func (c *Cache[K, T]) Get(key K) (T, error) {
	item, ok := c.cache[key]
	if !ok {
		var zeroValue T
		return zeroValue, ErrKeyNotFound(key)
	}
	return item, nil
}
