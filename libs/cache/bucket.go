package cache

import (
	"container/list"
	"sync"
	"time"
)

type bucket struct {
	items map[string]*item
	lock  sync.Mutex

	evictList *list.List
	maxLen    int

	//deleteChan      chan *list.Element
	//insertChan      chan *item
	//moveToFrontChan chan *item
	//evictChan       chan struct{}
}

type item struct {
	value     any
	expiredAt int64
	element   *list.Element
}

func initBucket(maxLen int) *bucket {
	b := &bucket{
		items:     make(map[string]*item, maxLen),
		lock:      sync.Mutex{},
		evictList: list.New(),
		maxLen:    maxLen,
		//deleteChan:      make(chan *list.Element),
		//insertChan:      make(chan *item),
		//moveToFrontChan: make(chan *item),
		//evictChan:       make(chan struct{}),
	}
	//go func() {
	//	for {
	//		select {
	//		case item, ok := <-b.deleteChan:
	//			if !ok {
	//				break
	//			}
	//			b.deleteFromEvictList(item)
	//		case item, ok := <-b.insertChan:
	//			if !ok {
	//				break
	//			}
	//			b.insertToEvictList(item)
	//		case item, ok := <-b.moveToFrontChan:
	//			if !ok {
	//				break
	//			}
	//			b.moveFrontInEvictList(item)
	//		case <-b.evictChan:
	//			b.onEvict()
	//		}
	//	}
	//}()
	return b
}

func (b *bucket) Get(key string) (interface{}, bool) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if val, ok := b.items[key]; ok {
		if time.Now().Unix() > val.expiredAt {
			return nil, false
		}
		b.evictList.MoveToFront(val.element)
		return val.value, ok
	} else {
		return nil, ok
	}
}
func (b *bucket) Set(key string, value interface{}, ttl time.Duration) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if val, ok := b.items[key]; ok {
		b.evictList.MoveToFront(val.element)
		return
	}
	if len(b.items) >= b.maxLen {
		item := b.evictList.Back()
		//logger.Debug("list", zap.Any("list", b.evictList.Len()), zap.Any("item", item))
		delete(b.items, item.Value.(string))
		b.evictList.Remove(item)
		//logger.Debug("evicted", zap.String("key", item.Value.(string)))
	}
	b.items[key] = &item{
		value:     value,
		expiredAt: time.Now().Add(ttl).Unix(),
		element:   b.evictList.PushFront(key),
	}
}

func (b *bucket) cleanup() {
	b.lock.Lock()
	keys := b.keys()
	b.lock.Unlock()
	for _, key := range keys {
		b.lock.Lock()
		item := b.items[key]
		if time.Now().Unix() > item.expiredAt {
			delete(b.items, key)
			b.evictList.Remove(item.element)
		}
		b.lock.Unlock()
	}
}

func (b *bucket) keys() []string {
	keys := make([]string, 0, len(b.items))
	for key, _ := range b.items {
		keys = append(keys, key)
	}
	return keys
}
