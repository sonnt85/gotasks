package gotasks

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// TestNew verifies that New returns a non-nil *T.
func TestNew(t *testing.T) {
	tch := New()
	if tch == nil {
		t.Fatal("New() returned nil")
	}
}

// TestGoAndWait_Success launches multiple tasks that succeed and verifies Wait returns nil.
func TestGoAndWait_Success(t *testing.T) {
	tch := New()

	var count atomic.Int32
	for i := 0; i < 5; i++ {
		tch.Go(func() error {
			count.Add(1)
			return nil
		})
	}

	if err := tch.Wait(); err != nil {
		t.Fatalf("Wait() returned unexpected error: %v", err)
	}

	if got := count.Load(); got != 5 {
		t.Fatalf("expected 5 tasks to run, got %d", got)
	}
}

// TestGoAndWait_Error launches tasks where one returns an error and verifies Wait returns it.
func TestGoAndWait_Error(t *testing.T) {
	tch := New()

	expectedErr := errors.New("task failed")

	tch.Go(func() error {
		return nil
	})
	tch.Go(func() error {
		return expectedErr
	})

	// Wait should return the error from the failing task.
	// Call Wait repeatedly until nil to drain all results.
	var gotErr error
	for {
		err := tch.Wait()
		if err != nil {
			gotErr = err
		}
		if err == nil {
			break
		}
	}

	if gotErr == nil {
		t.Fatal("Wait() should have returned an error, got nil")
	}
	if gotErr.Error() != expectedErr.Error() {
		t.Fatalf("Wait() returned %q, expected %q", gotErr.Error(), expectedErr.Error())
	}
}

// TestGoAndWait_MultipleErrors launches tasks with multiple errors and verifies
// Wait returns the first error received.
func TestGoAndWait_MultipleErrors(t *testing.T) {
	tch := New()

	err1 := errors.New("error one")
	err2 := errors.New("error two")

	tch.Go(func() error {
		return err1
	})
	tch.Go(func() error {
		return err2
	})
	tch.Go(func() error {
		return nil
	})

	// The first call to Wait returns whichever error is received first.
	firstErr := tch.Wait()
	if firstErr == nil {
		t.Fatal("first Wait() should have returned an error")
	}
	if firstErr.Error() != err1.Error() && firstErr.Error() != err2.Error() {
		t.Fatalf("first Wait() returned unexpected error: %v", firstErr)
	}

	// Drain remaining errors/completion by calling Wait again.
	for {
		err := tch.Wait()
		if err == nil {
			break
		}
		// Each subsequent error must be one of the known errors.
		if err.Error() != err1.Error() && err.Error() != err2.Error() {
			t.Fatalf("subsequent Wait() returned unexpected error: %v", err)
		}
	}
}

// TestGoAndWait_NoTasks verifies Wait returns nil when no tasks were launched.
func TestGoAndWait_NoTasks(t *testing.T) {
	tch := New()

	if err := tch.Wait(); err != nil {
		t.Fatalf("Wait() with no tasks returned unexpected error: %v", err)
	}
}

// TestGoAndWait_Concurrent verifies that tasks actually run concurrently.
// All tasks increment a counter atomically, then wait on a shared signal before
// completing. If tasks ran sequentially, the counter would never reach the total
// while any task is still waiting.
func TestGoAndWait_Concurrent(t *testing.T) {
	tch := New()

	const numTasks = 5
	var started atomic.Int32
	gate := make(chan struct{})

	for i := 0; i < numTasks; i++ {
		tch.Go(func() error {
			started.Add(1)
			// Block until all tasks have started.
			<-gate
			return nil
		})
	}

	// Poll until all tasks have started, proving concurrency.
	deadline := time.After(5 * time.Second)
	for {
		if started.Load() == numTasks {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for tasks to start concurrently; only %d/%d started", started.Load(), numTasks)
		default:
			time.Sleep(time.Millisecond)
		}
	}

	// Release all tasks.
	close(gate)

	if err := tch.Wait(); err != nil {
		t.Fatalf("Wait() returned unexpected error: %v", err)
	}
}
