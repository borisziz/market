package cache

import (
	"context"
	"hash/fnv"
	"time"
)

type cache struct {
	buckets []*bucket

	numBuckets      int32
	cleanupInterval time.Duration
	defaultTTL      time.Duration
}

func NewCache(numBuckets int32, defaultTTL time.Duration, cleanupInterval time.Duration, maxBucketLen int) *cache {
	if cleanupInterval == 0 {
		cleanupInterval = time.Minute
	}
	if defaultTTL == 0 {
		defaultTTL = time.Minute * 5
	}
	c := &cache{
		defaultTTL:      defaultTTL,
		cleanupInterval: cleanupInterval,
		numBuckets:      numBuckets,
		buckets:         make([]*bucket, 0, numBuckets),
	}
	for i := 0; i < int(numBuckets); i++ {
		c.buckets = append(c.buckets, initBucket(maxBucketLen))
	}

	ctx := context.Background()
	go func() {
		ticker := time.NewTicker(c.cleanupInterval)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				for _, b := range c.buckets {
					b.cleanup()
				}
			}
		}
	}()
	return c
}

/*
func NewCache1() {
	c, err := lru.New(1)
	l := list.New()
	g := l.PushBack()
	bigcache.BigCache{}

	cc := ccache.New()
}
*/

func (c *cache) Get(key string) (interface{}, bool) {
	h := fnv.New32a()
	h.Write([]byte(key))
	val, ok := c.buckets[int(h.Sum32())%int(c.numBuckets)].Get(key)
	return val, ok
}

func (c *cache) Set(key string, value interface{}, ttl time.Duration) {
	h := fnv.New32a()
	h.Write([]byte(key))
	c.buckets[int(h.Sum32())%int(c.numBuckets)].Set(key, value, ttl)
}
