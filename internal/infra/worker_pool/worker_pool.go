package workerpool

import (
	"log"
	"runtime"
	"sync"
)

type WorkerPool struct {
	workers  int
	taskChan chan func()
	wg       sync.WaitGroup
}

func New(maxWorkers int) *WorkerPool {
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}

	multiple := 2

	pool := &WorkerPool{
		workers:  maxWorkers,
		taskChan: make(chan func(), maxWorkers*multiple),
	}

	for range maxWorkers {
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

	p.wg.Add(1)
	p.taskChan <- task
}

func (p *WorkerPool) Wait() {
	p.wg.Wait()
}
