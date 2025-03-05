# sgoq
simple go queue - sgoq

## Simple asynchronous task queue

### To start:
```go run .```
![screen-record](https://github.com/clarkreiz/sgoq/blob/main/scr.gif)

# Description:

The core of the system consists of four main components:

- Queue – a simple asynchronous queue based on buffered channels and atomics. The queue is aware of tasks and their priorities, and it retrieves tasks based on priority.

- Worker – a simple worker, whose task is to constantly fetch tasks from the queue and execute them. He gets sad when he fetches an empty task. He only knows about Dequeue and IsStopped.

- WorkerPool – A tool for managing workers, it can increase or decrease the number of workers by a delta. It knows about the queue no more than the worker.

- Supervisor – the overseer of everything, who knows how many tasks are currently being worked on in the queue and can manage the number of workers via the WorkerPool. It contains mysterious logic for managing workers based on thresholds.
