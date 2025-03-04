package main

import (
	"context"
	"log"
	"sync"
)

type WorkerPool struct {
	ctx     context.Context
	cancel  context.CancelFunc
	queue   Queue
	workers map[int]chan struct{}
	mu      sync.Mutex
	wg      sync.WaitGroup
}

func NewWorkerPool(queue Queue, initialWorkers int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	pool := &WorkerPool{
		ctx:     ctx,
		cancel:  cancel,
		queue:   queue,
		workers: make(map[int]chan struct{}),
	}
	pool.Scale(initialWorkers)
	return pool
}

func (p *WorkerPool) Scale(delta int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	current := len(p.workers)
	target := current + delta
	if target < 0 {
		target = 0
	}

	if target > current {
		for i := current; i < target; i++ {
			stopChan := make(chan struct{})
			p.workers[i] = stopChan
			p.wg.Add(1)
			go func(id int, stopChan chan struct{}) {
				defer p.wg.Done()
				// TODO: сделать лучше
				worker := NewWorker(p.ctx, id, p.queue, stopChan)
				worker.Start()
			}(i, stopChan)
		}
	} else if target < current {
		for i := current - 1; i >= target; i-- {
			if stopChan, exists := p.workers[i]; exists {
				close(stopChan)
				delete(p.workers, i)
			}
		}
	}
}

func (p *WorkerPool) Current() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.workers)
}

func (p *WorkerPool) Shutdown() {
	log.Println("Starting graceful shutdown of worker pool...")

	p.cancel()

	// Wait for all workers to complete
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	<-done
	log.Println("Worker pool shutdown complete")
}
