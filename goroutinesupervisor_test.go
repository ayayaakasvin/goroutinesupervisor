package goroutinesupervisor

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTaskPanicRecovery(t *testing.T) {
    ctx := context.Background()
    gs := NewSupervisor(ctx)
    
    eventCaught := false
    gs.WithHandler(func(e Event) {
        if e.Type == EventTaskFailed && e.Panic {
            eventCaught = true
        }
    })
    
    gs.Go("panic-task", func(ctx context.Context) error {
        panic("test panic")
    })
    
    gs.Wait()
    assert.True(t, eventCaught)
}

func TestTaskExecution(t *testing.T) {
    ctx := context.Background()
    gs := NewSupervisor(ctx)
    
	const loops = 127
    counter := int32(0)
    
	for i := 1; i <= loops; i++ {
		gs.Go(fmt.Sprintf("Counter-Add-%d", i), func(ctx context.Context) error {
			atomic.AddInt32(&counter, 1)
			return nil
		})
	}
    
    gs.Wait()
    assert.True(t, loops == counter)
}

func TestTaskPartialFail(t *testing.T) {
    ctx := context.Background()
    gs := NewSupervisor(ctx)

    var counter int32
    var eventFail int32

    task2Start := make(chan struct{})

    gs.Go("running-task-1", func(ctx context.Context) error {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-time.After(time.Second * 10):
            atomic.AddInt32(&counter, 1)
            return nil
        }
    })

    gs.Go("failing-task-2", func(ctx context.Context) error {
        <-task2Start
        time.Sleep(50 * time.Millisecond)
        atomic.StoreInt32(&eventFail, 1)
        return errors.New("true error")
    })

    gs.Go("running-task-3", func(ctx context.Context) error {
        atomic.AddInt32(&counter, 1)
        close(task2Start)
        return nil
    })

    gs.Wait()

    assert.Equal(t, int32(1), atomic.LoadInt32(&eventFail))
    assert.Equal(t, int32(1), atomic.LoadInt32(&counter))
}