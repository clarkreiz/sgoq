package main

import (
	"context"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// priority level
const (
	critical = iota
	high
	medium
	low
	minimal

	numPriorities
)

const (
	qbuff          = 1000
	initialWorkers = 5
	minWorkers     = 1
	maxWorkers     = 100
)

func init() {
	log.SetFlags(log.Ltime)
	log.Println("Logger initialized")
}

func main() {
	log.Println("Start....")

	pq := NewPriorityQueue(qbuff)
	pool := NewWorkerPool(pq, initialWorkers)
	supervisor := NewSupervisor(pool, pq, minWorkers, maxWorkers)

	supervisor.Start()
	
	// To get classic log output, simply comment out the line below.
	go report(pq, pool)

	// Simulate infinite task addition
	go func() {
		for {
			for i := 0; i < 1000; i++ {
				err := pq.Enqueue(makeTask())
				if err != nil {
					// if queue is full, wait a bit
					time.Sleep(time.Second)
				}
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("\nReceived signal: %v\n", sig)

	shutdown(pq, pool, supervisor)
}

func shutdown(pq *PriorityQueue, pool *WorkerPool, supervisor *Supervisor) {
	log.Println("Initiating graceful shutdown...")

	pq.Stop()
	log.Println("Stopped accepting new tasks")

	supervisor.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start shutdown in goroutine
	done := make(chan struct{})
	go func() {
		pool.Shutdown()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Shutdown complete")
	case <-ctx.Done():
		log.Println("Shutdown timed out after 30 seconds")
	}
}

func report(pq *PriorityQueue, pool *WorkerPool) {
	priorityNames := []string{
		"CRITICAL",
		"HIGH",
		"MEDIUM",
		"LOW",
		"MIN",
	}

	for {
		log.Print("\033[H\033[2J")

		log.Println("SYSTEM STATUS")
		log.Println("-------------")
		log.Printf("Active Workers: %d\n", pool.Current())
		log.Println()
		log.Println("PRIORITY   TASKS COUNT")
		log.Println("---------------------")

		for i := 0; i < numPriorities; i++ {
			count := int(pq.taskByPrior[i].Load())
			log.Printf("%-10s %d\n", priorityNames[i], count)
		}

		time.Sleep(500 * time.Millisecond)
	}
}

func makeTask() *Task {
	randomPrior := rand.Intn(numPriorities)
	randomID := rand.Intn(1e6)
	return &Task{
		ID:        randomID,
		Priority:  randomPrior,
		Payload:   struct{}{},
		CreatedAt: time.Now(),
		Exe: func() {
			duration := time.Duration(rand.Intn(900)+100) * time.Millisecond
			time.Sleep(duration)
		},
	}
}
