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
	}
	return b
}

func (b *bucket) Get(key string) (interface{}, bool) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if val, ok := b.items[key]; ok {
		if time.Now().Unix() > val.expiredAt {
			//Если элемент просрочен, но чистка еще не прошла, не будем его возвращать
			return nil, false
		}
		//Если все нашлось, то подвинем этот элемент в начало списка, как самый недавно запрошенный
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
		//Если элемент уже есть просто двигаем его вверх в списке
		b.evictList.MoveToFront(val.element)
		return
	}
	if len(b.items) >= b.maxLen {
		//Если достигли максимума значений в бакете, то удаляем тот, который запрашивали давнее всех
		item := b.evictList.Back()
		delete(b.items, item.Value.(string))
		b.evictList.Remove(item)
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
