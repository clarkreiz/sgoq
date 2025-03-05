package main

import (
	"log"
	"runtime"
	"sync"
	"time"
)

type WorkerPoolManager interface {
	Scale(delta int)
	Current() int
}

type QueueManager interface {
	GetCapacity() int
	GetTotalTasks() int
}

type Supervisor struct {
	pool               WorkerPoolManager
	queue              QueueManager
	minWorkers         int
	maxWorkers         int
	scaleUpThreshold   float64
	scaleDownThreshold float64
	scaleUpFactor      float64
	scaleDownFactor    float64

	mu       sync.Mutex
	stopChan chan struct{}
}

func NewSupervisor(pool WorkerPoolManager, queue QueueManager, minWorkers, maxWorkers int) *Supervisor {
	return &Supervisor{
		pool:               pool,
		queue:              queue,
		minWorkers:         minWorkers,
		maxWorkers:         maxWorkers,
		scaleUpThreshold:   0.7,
		scaleDownThreshold: 0.3,
		scaleUpFactor:      1.2,
		scaleDownFactor:    0.8,
		stopChan:           make(chan struct{}),
	}
}

func (s *Supervisor) Start() {
	go s.monitor()
}

func (s *Supervisor) Stop() {
	close(s.stopChan)
}

func (s *Supervisor) monitor() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.adjustWorkers()
		}
	}
}

// Working with thresholds is never easy, but I wrote a test and everything seems to work correctly.
func (s *Supervisor) adjustWorkers() {
	// TODO: recover from panic may be here
	s.mu.Lock()
	defer s.mu.Unlock()
	totalTasks := s.queue.GetTotalTasks()
	queueCapacity := s.queue.GetCapacity()
	if queueCapacity == 0 {
		log.Println("Queue capacity is zero, skipping scaling")
		return
	}

	util := float64(totalTasks) / float64(queueCapacity)
	currentWorkers := s.pool.Current()

	log.Printf("Runtime goroutines: %v, Workers: %d, Util: %.2f%%",
		runtime.NumGoroutine(), currentWorkers, util*100)

	if util > s.scaleUpThreshold && currentWorkers < s.maxWorkers {
		newWorkers := min(s.maxWorkers, int(float64(currentWorkers)*s.scaleUpFactor))
		if newWorkers > currentWorkers {
			log.Printf("Scaling up: %d -> %d", currentWorkers, newWorkers)
			s.pool.Scale(newWorkers - currentWorkers)
		}
	} else if util < s.scaleDownThreshold && currentWorkers > s.minWorkers {
		newWorkers := max(s.minWorkers, int(float64(currentWorkers)*s.scaleDownFactor))
		if newWorkers < currentWorkers {
			log.Printf("Scaling down: %d -> %d", currentWorkers, newWorkers)
			s.pool.Scale(newWorkers - currentWorkers)
		}
	}
}
