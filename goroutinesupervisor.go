// Package goroutinesupervisor provides a lightweight supervisor for managing
// the lifecycle of goroutine tasks. It supports named tasks, panic recovery,
// event emission and graceful cancellation via context.
package goroutinesupervisor

import (
	"context"
	"fmt"
	"os"
	"sync"
)

// GoRoutineSupervisor supervises goroutine tasks started via Go or
// GoInterrupt. It tracks running tasks, broadcasts lifecycle events to
// registered handlers, and exposes a context that is cancelled on first
// failure.
type GoRoutineSupervisor struct {
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	emitWg  sync.WaitGroup
	rwMutex sync.RWMutex
	errCh   chan error

	handlers []EventHandler
}

// NewSupervisor creates a supervisor rooted at the provided parent context.
// The returned supervisor exposes a Context() that will be cancelled on the
// first task failure or when Cancel is invoked via the context.
func NewSupervisor(parent context.Context) *GoRoutineSupervisor {
	ctx, cancel := context.WithCancel(parent)

	return &GoRoutineSupervisor{
		ctx:    ctx,
		cancel: cancel,
		errCh:  make(chan error, 1),
	}
}

// Go starts a named task under supervision. The task should observe the
// provided context to support cancellation. Any returned error or panic will
// cause the supervisor to cancel its context and emit a failure event.
func (g *GoRoutineSupervisor) Go(name string, task TaskRunner) {
	g.runTask(name, task, nil)
}

// GoInterrupt is like Go but accepts an interrupter callback that will be
// invoked with the error when a task fails.
func (g *GoRoutineSupervisor) GoInterrupt(name string, task TaskRunner, intr Interrupter) {
	g.runTask(name, task, intr)
}

// Wait blocks until all supervised tasks and pending event handlers complete.
// It returns the first observed error if any task failed.
func (g *GoRoutineSupervisor) Wait() error {
	g.wg.Wait()
	g.emitWg.Wait()

	select {
	case err := <-g.errCh:
		return err
	default:
		return nil
	}
}

// Context returns the supervisor's context which is cancelled on first failure.
func (g *GoRoutineSupervisor) Context() context.Context {
	return g.ctx
}

// WithHandler registers an event handler to receive task lifecycle events.
// Handlers are invoked asynchronously and panics from handlers are recovered.
func (g *GoRoutineSupervisor) WithHandler(h EventHandler) {
	g.rwMutex.Lock()
	g.handlers = append(g.handlers, h)
	g.rwMutex.Unlock()
}

// emit sends an Event to all registered handlers asynchronously. Handler
// panics are recovered and logged to stderr.
func (g *GoRoutineSupervisor) emit(e Event) {
	g.emitWg.Add(1)
	go func() {
		defer g.emitWg.Done()
		g.rwMutex.RLock()
		handlers := append([]EventHandler(nil), g.handlers...)
		g.rwMutex.RUnlock()

		for _, h := range handlers {
			func(h EventHandler) {
				defer func() {
					if r := recover(); r != nil {
						fmt.Fprintf(os.Stderr, "handler panic: %v\n", r)
					}
				}()

				h(e)
			}(h)
		}
	}()
}
