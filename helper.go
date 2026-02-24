package goroutinesupervisor

import (
	"fmt"
	"time"
)

// Helper functions to seperate parts of executable code and the boilerplate

func (g *GoRoutineSupervisor) runTask(name string, task TaskRunner, intr Interrupter) {
    g.wg.Add(1)

    go func() {
        start := time.Now()

        defer g.wg.Done()

        defer func() {
            if r := recover(); r != nil {
                err := fmt.Errorf("panic: %v", r)

                g.handleFailure(name, err, start, true, intr)
            }
        }()

        g.emit(Event{
            Type:    EventTaskStarted,
            Task:    name,
            Started: start,
        })

        if err := task(g.ctx); err != nil {
            g.handleFailure(name, err, start, false, intr)
            return
        }

        g.emit(Event{
            Type:    EventTaskFinished,
            Task:    name,
            Started: start,
            Ended:   time.Now(),
        })
    }()
}

func (g *GoRoutineSupervisor) handleFailure(
    name string,
    err error,
    start time.Time,
    panic bool,
    intr Interrupter,
) {
    end := now()

    g.emit(Event{
        Type:    EventTaskFailed,
        Task:    name,
        Err:     err,
        Panic:   panic,
        Started: start,
        Ended:   end,
    })

    select {
    case g.errCh <- err:
    default:
    }

    if intr != nil {
        intr(err)
    }

    g.cancel()
}

func now () time.Time {
	return time.Now()
}