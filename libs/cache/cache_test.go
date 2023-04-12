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

func BenchmarkMyCache(b *testing.B) {
	myCache := NewCache(11, 180*time.Second, 6*time.Minute, 1000)
	goCache := gocache.New(1*time.Minute, 5*time.Minute)
	bigCache, _ := bigcache.New(context.Background(), bigcache.DefaultConfig(10*time.Minute))
	lruCache, _ := hashicorp.New(10000)
	hits := 0
	misses := 0

	b.ResetTimer()

	b.Run("Set_my_cache", func(b *testing.B) {
		if b.N == 1 {
			return
		}
		alloc1 := memAlloc()
		for i := 0; i < b.N; i++ {
			myCache.Set(fmt.Sprintf("%d", i), gofakeit.Animal(), 15*time.Second)
		}
		alloc2 := memAlloc()
		b.ReportMetric(float64(alloc2-alloc1), "allocated_bytes_total")
	})

	b.Run("Add_go_cache", func(b *testing.B) {
		if b.N == 1 {
			return
		}
		alloc1 := memAlloc()
		for i := 0; i < b.N; i++ {
			goCache.Add(fmt.Sprintf("%d", i), gofakeit.Animal(), gocache.DefaultExpiration)
		}
		alloc2 := memAlloc()
		b.ReportMetric(float64(alloc2-alloc1), "allocated_bytes_total")
	})
	b.Run("Set_big_cache", func(b *testing.B) {
		if b.N == 1 {
			return
		}
		alloc1 := memAlloc()
		for i := 0; i < b.N; i++ {
			bigCache.Set(fmt.Sprintf("%d", i), []byte(fmt.Sprintf("%v", gofakeit.Animal())))
		}
		alloc2 := memAlloc()
		b.ReportMetric(float64(alloc2-alloc1), "allocated_bytes_total")
	})
	b.Run("Add_lru_cache", func(b *testing.B) {
		if b.N == 1 {
			return
		}
		alloc1 := memAlloc()
		for i := 0; i < b.N; i++ {
			lruCache.Add(fmt.Sprintf("%d", i), gofakeit.Animal())
		}
		alloc2 := memAlloc()
		b.ReportMetric(float64(alloc2-alloc1), "allocated_bytes_total")
	})

	b.Run("Get_my_cache", func(b *testing.B) {
		if b.N == 1 {
			return
		}
		for i := 0; i < b.N; i++ {
			value, _ := myCache.Get(fmt.Sprintf("%d", i))
			if value != nil {
				hits++
				_ = value
			} else {
				misses++
			}
		}
		b.ReportMetric(float64(hits/b.N), "hits_rate")
	})
	hits = 0
	misses = 0
	b.Run("Get_go_cache", func(b *testing.B) {
		if b.N == 1 {
			return
		}
		for i := 0; i < b.N; i++ {
			value, found := goCache.Get(fmt.Sprintf("%d", i))
			if found {
				hits++
				_ = value
			} else {
				misses++
			}
		}
		b.ReportMetric(float64(hits/b.N), "hits_rate")

	})
	hits = 0
	misses = 0
	b.Run("Get_big_cache", func(b *testing.B) {
		if b.N == 1 {
			return
		}
		for i := 0; i < b.N; i++ {
			value, _ := bigCache.Get(fmt.Sprintf("%d", i))
			if value != nil {
				hits++
				_ = value
			} else {
				misses++
			}
		}
		b.ReportMetric(float64(hits/b.N), "hits_rate")
	})
	hits = 0
	misses = 0
	b.Run("Get_lru_cache", func(b *testing.B) {
		if b.N == 1 {
			return
		}
		for i := 0; i < b.N; i++ {

			value, err := lruCache.Get(fmt.Sprintf("%d", i))
			if err == true {
				hits++
				_ = value
			} else {
				misses++
			}
		}
		b.ReportMetric(float64(hits/b.N), "hits_rate")

	})
}

func memAlloc() uint64 {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}
