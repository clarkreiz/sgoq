package main

import (
	"log"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// priority level
const (
	critical = iota
	high
	medium
	low
	min

	numPriorities
)

func init() {
	log.SetFlags(log.Ltime)
	log.Println("Logger initialized")
}

func main() {
	log.Println("Start....")
	concurrency := 70
	qbuff := 1000
	var wg sync.WaitGroup

	pq := NewPriorityQueue(qbuff)
	worker := NewWorker(pq, concurrency)

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Simulate infinit task addition
	go func() {
		for {
			for i := 0; i < 1000; i++ {
				err := pq.Enqueue(makeTask())
				if err != nil {
					// if queue is full, wait second
					time.Sleep(time.Second)
				}
				time.Sleep(time.Microsecond * 50)
			}
		}
	}()

	go printCountOfTask(pq)

	worker.StartProcessing(&wg)

	sig := <-sigChan
	log.Printf("\nReceived signal: %v\n", sig)

	worker.Shutdown(&wg, 30*time.Second)
}

func printCountOfTask(pq *PriorityQueue) {
	priorityNames := []string{
		"CRITICAL",
		"HIGH",
		"MEDIUM",
		"LOW",
		"MIN",
	}

	for {
		log.Print("\033[H\033[2J")

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
