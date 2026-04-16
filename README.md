# gotasks

Minimal concurrent task runner — launch goroutines and wait for the first error or all to succeed.

## Installation

```bash
go get github.com/sonnt85/gotasks
```

## Features

- Launch multiple goroutines as "tasks" with a single API
- `Wait()` blocks until all tasks succeed or any task fails
- Returns the first error encountered; call `Wait()` repeatedly to drain remaining tasks
- Zero dependencies — standard library only

## Usage

```go
import "github.com/sonnt85/gotasks"

t := gotasks.New()

t.Go(func() error {
    // task 1
    return doWork()
})

t.Go(func() error {
    // task 2
    return doOtherWork()
})

// Wait for all tasks to finish or any to fail
if err := t.Wait(); err != nil {
    log.Fatal(err)
}
```

## API

- `New() *T` — create a new task manager
- `(*T).Go(task func() error)` — launch a task as a goroutine
- `(*T).Wait() error` — block until all tasks finish or one returns an error; returns `nil` when all succeed

## License

MIT License - see [LICENSE](LICENSE) for details.
