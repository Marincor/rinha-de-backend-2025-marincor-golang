package workerpool

import (
	"log"
	"runtime"
	"sync"
)

type WorkerPool struct {
	workers    int
	taskChan   chan func()
	wg         sync.WaitGroup
	closed     bool
	closeMutex sync.RWMutex
}

func New(maxWorkers int) *WorkerPool {
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}

	pool := &WorkerPool{
		workers:  maxWorkers,
		taskChan: make(chan func(), maxWorkers*2), // buffer para reduzir blocking
	}

	// Inicia os workers imediatamente
	for i := 0; i < maxWorkers; i++ {
		go pool.worker()
	}

	return pool
}

func (p *WorkerPool) worker() {
	for task := range p.taskChan {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Print(
						map[string]interface{}{
							"message": "error executing task",
							"error":   r,
						},
					)
				}
				p.wg.Done()
			}()
			task()
		}()
	}
}

func (p *WorkerPool) Submit(task func()) {
	if task == nil {
		return
	}

	p.closeMutex.RLock()
	defer p.closeMutex.RUnlock()

	if p.closed {
		return
	}

	p.wg.Add(1)
	p.taskChan <- task
}

func (p *WorkerPool) Wait() {
	p.wg.Wait()
}
