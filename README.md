# GoRoutine Supervisor

Lightweight supervisor for managing goroutine lifecycles with graceful shutdown, panic recovery, event emission and optional interruption callbacks.

Features:
- Start named tasks with NewSupervisor(ctx).Go(name, task) or GoInterrupt(name, task, intr).
- Automatic panic recovery and error propagation.
- Event handlers via WithHandler(func(Event)) for TaskStarted, TaskFinished, TaskFailed.
- Wait() blocks until all tasks and event handlers finish; returns first error if any.
- Access supervisor context with Context() for cancellation-aware tasks.

Quick use:
1. s := NewSupervisor(context.Background())
2. s.Go("worker", func(ctx context.Context) error { /* work */ return nil })
3. s.WithHandler(func(e Event) { /* monitor events */ })
4. err := s.Wait()

See goroutinesupervisor.go and other source files for implementation and tests.