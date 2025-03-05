package main

import (
	"context"
	"log"
)

type Queue interface {
	Dequeue() *Task
	IsStopped() bool
}

type Worker struct {
	id       int
	ctx      context.Context
	queue    Queue
	stopChan chan struct{}
}

func NewWorker(ctx context.Context, id int, queue Queue, stopChan chan struct{}) *Worker {
	return &Worker{
		id:       id,
		ctx:      ctx,
		queue:    queue,
		stopChan: stopChan,
	}
}

func (w *Worker) Start() {
	for {
		select {
		case <-w.ctx.Done():
			log.Printf("Worker %d stopping", w.id)
			return
		case <-w.stopChan:
			log.Printf("Worker %d received stop signal", w.id)
			return
		default:
			if task := w.queue.Dequeue(); task != nil {
				task.Exe()
			} else if w.queue.IsStopped() {
				return
			}
			// I haven't figured out what to do if the task hasn't arrived yet, 
			// the worker will probably be sad
		}
	}
}
