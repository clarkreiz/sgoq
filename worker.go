package main

import (
	"context"
	"log"
	"sync"
	"time"
)

type Queue interface {
	Enqueue(*Task) error
	Dequeue() *Task
	Close()
	IsEmpty() bool
	Stop()
	IsStopped() bool
}

type Worker struct {
	pq          Queue
	concurrency int
	ctx         context.Context
	cancel      context.CancelFunc
}

// Shutdown mark queue as stopped to prevent new tasks and
// wait for workers to finish with timeout.
func (w *Worker) Shutdown(wg *sync.WaitGroup, timeout time.Duration) {
	log.Println("Starting graceful shutdown...")

	w.pq.Stop()
	log.Println("Stopped accepting new jobs")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), timeout)
	defer shutdownCancel()

	w.cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All workers completed gracefully")
	case <-shutdownCtx.Done():
		log.Println("Shutdown timed out after", timeout)
	}

	log.Println("Worker shutdown complete")
}

func (w *Worker) StartProcessing(wg *sync.WaitGroup) {
	for i := 0; i < w.concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-w.ctx.Done():
					log.Printf("Worker %d is stopping\n", workerID)
					return
				default:
					task := w.pq.Dequeue()
					if task != nil {
						// TODO: capture panic here
						task.Exe()
					} else {
						// If queue is empty and stopped, exit
						if w.pq.IsStopped() {
							return
						}
						// Otherwise wait a bit and try again
						time.Sleep(time.Millisecond * 100)
					}
				}
			}
		}(i)
	}
}

func NewWorker(pq Queue, concurrency int) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{
		pq:          pq,
		concurrency: concurrency,
		ctx:         ctx,
		cancel:      cancel,
	}
}
