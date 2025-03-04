# sgoq
simple go queue - sgoq

## Simple asynchronous task queue

### To start the visualization of the queue:
```go run .```

![screen-record](https://github.com/clarkreiz/sgoq/blob/main/screen3.gif)

Core requirements:

- ~~Spawn N worker goroutines that process tasks from a shared queue~~
- ~~Limit buffer size to **1000** tasks~~
- ~~Implement priority-based task execution (1-5, 1=highest)~~
- ~~Handle SIGTERM/SIGINT to:~~
    - ~~Stop accepting new jobs~~
    - ~~Complete in-progress jobs~~
    - ~~Timeout after 30 seconds for pending jobs~~
