package cache

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"hash/fnv"
	"time"
)

type cache struct {
	buckets []*bucket

	numBuckets      int32
	cleanupInterval time.Duration
	defaultTTL      time.Duration
}

var (
	RequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "homework",
		Subsystem: "cache",
		Name:      "requests_total",
	})
	HistogramResponseTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "homework",
		Subsystem: "cache",
		Name:      "histogram_response_time_seconds",
		Buckets:   prometheus.ExponentialBuckets(0.0001, 2, 16),
	},
		[]string{"hit"},
	)
)

func NewCache(numBuckets int32, defaultTTL time.Duration, cleanupInterval time.Duration, maxBucketLen int) *cache {
	if cleanupInterval == 0 {
		cleanupInterval = time.Minute
	}
	if defaultTTL == 0 {
		defaultTTL = -1
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
				//Раз в заданный интервал времени в каждом бакете удаляются просроченные значения
				for _, b := range c.buckets {
					b.cleanup()
				}
			}
		}
	}()
	return c
}

func (c *cache) Get(key string) (interface{}, bool) {
	RequestsTotal.Inc()
	timeStart := time.Now()
	h := fnv.New32a()
	h.Write([]byte(key))
	//По остатку от деления на количество бакетов значение берется из определенного
	val, ok := c.buckets[int(h.Sum32())%int(c.numBuckets)].Get(key)
	defer HistogramResponseTime.WithLabelValues(fmt.Sprintf("%v", ok)).Observe(time.Since(timeStart).Seconds())
	return val, ok
}

func (c *cache) Set(key string, value interface{}, ttl time.Duration) {
	if ttl == 0 {
		ttl = c.defaultTTL
	}
	h := fnv.New32a()
	h.Write([]byte(key))
	//По остатку от деления хэша ключа на количество бакетов значение кладется в определенный бакет
	c.buckets[int(h.Sum32())%int(c.numBuckets)].Set(key, value, ttl)
}
