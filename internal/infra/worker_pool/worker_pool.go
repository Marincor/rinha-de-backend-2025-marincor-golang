package workerpool

import (
	"sync"
)

type WorkerPool struct {
	semaphore chan struct{}
	wg        sync.WaitGroup
}

func New(maxWorkers int) *WorkerPool {
	return &WorkerPool{
		semaphore: make(chan struct{}, maxWorkers),
	}
}

func (p *WorkerPool) Submit(callback func()) {
	p.wg.Add(1)
	p.semaphore <- struct{}{} // get slot

	go func() {
		defer func() {
			<-p.semaphore // release slot
			p.wg.Done()
		}()
		callback()
	}()
}

func (p *WorkerPool) Wait() {
	p.wg.Wait()
}
