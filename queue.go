package main

import (
	"errors"
	"fmt"
	"log"
	"sync/atomic"
	"time"
)

type Task struct {
	ID        int
	Priority  int
	Payload   any
	CreatedAt time.Time
	Exe       func()
	executed  atomic.Bool // for test
}

const cpuSmoko = time.Millisecond * 100

// Priority queue with 5 queues and an additional priority level.
// Workers pull tasks from the queues, starting with the critical priority.
// In this case, we can use atomic because there are simple inc/dec
// no nested operations, just chanel and atomic
type PriorityQueue struct {
	queues      [numPriorities]chan *Task
	taskByPrior [numPriorities]atomic.Int32
	totalTask   atomic.Int32
	stopped     atomic.Bool
	capacity    int
}

func (pq *PriorityQueue) IsEmpty() bool {
	return pq.totalTask.Load() == 0
}

func (pq *PriorityQueue) Close() {
	log.Println("PriorityQueue: closing all queues")
	for i := range pq.queues {
		close(pq.queues[i])
	}
}

func (pq *PriorityQueue) Stop() {
	log.Println("PriorityQueue: stopping task processing")
	pq.stopped.Store(true)
}

func (pq *PriorityQueue) IsStopped() bool {
	return pq.stopped.Load()
}

// Enqueue enqueue task if queue is not stopped and not full.
func (pq *PriorityQueue) Enqueue(task *Task) error {
	if pq.IsStopped() {
		log.Printf("PriorityQueue: rejected task %d, queue is stopped", task.ID)
		return errors.New("queue is stopped")
	}
	select {
	case pq.queues[task.Priority] <- task:
		pq.totalTask.Add(1)
		pq.taskByPrior[task.Priority].Add(1)
		return nil
	default:
		// Queue is full, give the CPU a small break and throw an error.

		// imagine that we backup our task and incremt the metric here,
		// We will attach an alert to our metric
		// that will be triggered to avoid disk or memory overflow. So goood...
		time.Sleep(cpuSmoko)
		return fmt.Errorf("Queue is full for priority %d", task.Priority)
	}
}

// Dequeue: try to pull of task ordered by priority
// in prod should be better to use a weights for each q
func (pq *PriorityQueue) Dequeue() *Task {
	for i := critical; i < numPriorities; i++ {
		select {
		case task, ok := <-pq.queues[i]:
			if !ok {
				continue
			}
			pq.totalTask.Add(-1)
			pq.taskByPrior[i].Add(-1)
			return task
		default:
			// Try next priority level
			time.Sleep(cpuSmoko)
		}
	}
	return nil
}

func (pq *PriorityQueue) GetCapacity() int {
	return pq.capacity
}

func (pq *PriorityQueue) GetTotalTasks() int {
	return int(pq.totalTask.Load())
}

func NewPriorityQueue(size int) *PriorityQueue {
	pq := &PriorityQueue{capacity: size}

	for i := critical; i < len(pq.queues); i++ {
		pq.queues[i] = make(chan *Task, size/numPriorities)
	}

	return pq
}
