package cache

import (
	"context"
	"fmt"
	"github.com/allegro/bigcache/v3"
	"github.com/brianvoe/gofakeit/v6"
	hashicorp "github.com/hashicorp/golang-lru"
	gocache "github.com/patrickmn/go-cache"
	"runtime"
	"testing"
	"time"
)

func BenchmarkCache(b *testing.B) {
	c := NewCache(11, 180*time.Second, 6*time.Minute, 1000)
	hits := 0
	misses := 0
	b.ResetTimer()
	b.Run("Set", func(b *testing.B) {
		alloc1 := memAlloc()
		for i := 0; i < b.N; i++ {
			c.Set(fmt.Sprintf("%d", i), gofakeit.Animal(), 15*time.Second)
		}
		alloc2 := memAlloc()
		b.Log("bytes", alloc2-alloc1)
	})
	b.Run("Get", func(b *testing.B) {
		alloc1 := memAlloc()
		for i := 0; i < b.N; i++ {
			value, _ := c.Get(fmt.Sprintf("%d", i))
			if value != nil {
				hits++
				_ = value
			} else {
				misses++
			}
		}
		alloc2 := memAlloc()
		b.Log("bytes", alloc2-alloc1)
		b.Log("hits", hits, "misses", misses)
	})
}

func BenchmarkGoCache(b *testing.B) {
	c := gocache.New(1*time.Minute, 5*time.Minute)
	hits := 0
	misses := 0
	b.ResetTimer()
	b.Run("Add", func(b *testing.B) {
		alloc1 := memAlloc()
		for i := 0; i < b.N; i++ {
			c.Add(fmt.Sprintf("%d", i), gofakeit.Animal(), gocache.DefaultExpiration)
		}
		alloc2 := memAlloc()
		b.Log("bytes", alloc2-alloc1)
	})
	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			value, found := c.Get(fmt.Sprintf("%d", i))
			if found {
				hits++
				_ = value
			} else {
				misses++
			}
		}
		b.Log("hits", hits, "misses", misses)
	})
}

func BenchmarkBigCache(b *testing.B) {
	c, _ := bigcache.New(context.Background(), bigcache.DefaultConfig(10*time.Minute))
	hits := 0
	misses := 0
	b.ResetTimer()
	b.Run("Set", func(b *testing.B) {
		alloc1 := memAlloc()
		for i := 0; i < b.N; i++ {
			c.Set(fmt.Sprintf("%d", i), []byte(fmt.Sprintf("%v", gofakeit.Animal())))
		}
		alloc2 := memAlloc()
		b.Log("bytes", alloc2-alloc1)
	})
	b.Run("Get2", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			value, _ := c.Get(fmt.Sprintf("%d", i))
			if value != nil {
				hits++
				_ = value
			} else {
				misses++
			}
		}
		b.Log("hits", hits, "misses", misses)
	})
}

func BenchmarkCacheLRU(b *testing.B) {
	ccc, _ := hashicorp.New(10000)
	misses := 0
	hits := 0
	b.ResetTimer()

	b.Run("Add", func(b *testing.B) {
		alloc1 := memAlloc()
		for i := 0; i < b.N; i++ {
			ccc.Add(fmt.Sprintf("%d", i), gofakeit.Animal())
		}
		alloc2 := memAlloc()
		b.Log("bytes", alloc2-alloc1)
	})

	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {

			value, err := ccc.Get(fmt.Sprintf("%d", i))
			if err == true {
				hits++
				_ = value
			} else {
				misses++
			}
		}
		b.Log("hits", hits, "misses", misses)

	})

}

func memAlloc() uint64 {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}
