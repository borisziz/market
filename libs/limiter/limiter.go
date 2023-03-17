package limiter

import (
	"context"
	"time"
)

type Limiter interface {
	Wait(ctx context.Context) error
}

var _ Limiter = (*limiter)(nil)

type limiter struct {
	wait  chan struct{}
	limit time.Duration
	burst int32
}

func NewLimiter(countPerSecond, burst int32) Limiter {
	l := &limiter{
		limit: time.Second / time.Duration(countPerSecond),
		burst: burst,
		//burst дает возможность буферезировать канал, и тем самым прочитать оттуда сразу n значений, в случае высокой нагрузки
		wait: make(chan struct{}, burst),
	}
	go l.run()
	return l
}

func (l *limiter) run() {
	//Блокируемся, пока не будет вызван Wait
	for range time.Tick(l.limit) {
		l.wait <- struct{}{}
	}
}

func (l *limiter) Wait(ctx context.Context) error {
	//При отмене контекста выходит не дожидаясь
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-l.wait:
			return nil
		}
	}
}
