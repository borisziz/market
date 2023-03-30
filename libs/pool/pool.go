package pool

//go:generate sh -c "rm -rf mocks && mkdir -p mocks"
//go:generate minimock -i Pool -o ./mocks/ -s "_minimock.go"

import (
	"context"
	"sync"
)

type Task struct {
	//Создание отдано на откуп клиенту
	//Результаты на стороне клиента можно как складывать в канал, так и сразу присваивать куда-то
	//А ошибки буду возвращаться через канал
	Task func() error
	//Количество попыток
	attempts uint8
}

type Pool interface {
	Submit(Task)
	Close()
}

var _ Pool = (*pool)(nil)

type pool struct {
	ctx           context.Context
	cancel        context.CancelFunc
	amountWorkers uint16
	//Чтобы в конце ждать пока все воркеры закончат свою работу
	wgWorkers sync.WaitGroup
	//Чтобы прежде чем закрыть канал с тасками, убеждаться что все таски закончены(по скольку могут быть повторные попытки)
	wgTasks sync.WaitGroup
	//Количество повторных попыток
	maxRetries uint8

	tasks  chan Task
	errors chan error
}

func NewPool(ctx context.Context, amountWorkers uint16, maxRetries uint8, withCancelOnError bool) (Pool, <-chan error) {
	var cancel context.CancelFunc
	if withCancelOnError {
		ctx, cancel = context.WithCancel(ctx)
	}
	pool := &pool{
		ctx:           ctx,
		cancel:        cancel,
		amountWorkers: amountWorkers,
		tasks:         make(chan Task, amountWorkers),
		errors:        make(chan error, amountWorkers),
		maxRetries:    maxRetries,
	}
	pool.startWorkers()
	return pool, pool.errors
}

func (p *pool) startWorkers() {
	p.wgWorkers.Add(int(p.amountWorkers))
	for i := 0; i < int(p.amountWorkers); i++ {
		go func() {
			defer p.wgWorkers.Done()
			worker(p.ctx, p.tasks, p.execute)
		}()
	}
}

func (p *pool) Submit(task Task) {
	//Пока контекст не завершен, отправляем таски
	if task.attempts == 0 {
		p.wgTasks.Add(1)
	}
	select {
	case <-p.ctx.Done():
		return
	case p.tasks <- task:
	}
}

func (p *pool) Close() {
	defer p.cancel()
	p.wgTasks.Wait()
	//Закрываем канал с тасками и ждем пока отработают воркеры
	close(p.tasks)
	p.wgWorkers.Wait()
	close(p.errors)
}

func (p *pool) execute(task Task) {
	err := task.Task()
	if err != nil {
		if task.attempts+1 < p.maxRetries {
			//Отправляем ту же таску исполняться (в фоне, чтобы не блокировать worker-а) еще раз инкрементируя счетчик попыток
			task.attempts++
			go p.Submit(task)
		} else {
			defer p.wgTasks.Done()
			p.errors <- err
			if p.cancel != nil {
				p.cancel()
			}
		}
	} else {
		p.wgTasks.Done()
	}
}

func worker(
	ctx context.Context,
	tasks <-chan Task,
	executor func(Task),
) {
	for {
		select {
		case <-ctx.Done():
			return
		case task, ok := <-tasks:
			if task.Task == nil || !ok {
				return
			}
			//Отправляем в функцию, которая имеет доступ к полям пула
			executor(task)
		}
	}
}
