package goshutdownchannel

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
)

type Shutdown struct {
	ctx      context.Context
	cancel   context.CancelFunc
	msgCh    chan string
	signalCh chan os.Signal
	mu       sync.RWMutex
	lastMsg  string
}

// NewShutdown creates a new Shutdown instance with a cancellable context.
func NewShutdown(ctx context.Context, cancel context.CancelFunc) *Shutdown {
	return &Shutdown{
		ctx:      ctx,
		cancel:   cancel,
		msgCh:    make(chan string, 1),
		signalCh: nil,
	}
}

// Send sends a message to the message channel and updates the last message.
func (s *Shutdown) Send(origin, msg string) {
	full := fmt.Sprintf("%s:%s", origin, msg)

	// store last message safely
	s.mu.Lock()
	s.lastMsg = full
	s.mu.Unlock()

	// keep non-blocking send to channel (buffered)
	select {
	case s.msgCh <- full:
	default:
	}
	s.cancel()
}

// Done returns a channel that is closed when the context is done.
func (s *Shutdown) Done() <-chan struct{} {
	return s.ctx.Done()
}

// Message returns the most recently sent message (non-destructive, safe).
func (s *Shutdown) Message() string {
	s.mu.RLock()
	m := s.lastMsg
	s.mu.RUnlock()
	return m
}

// Notify sets up a signal channel to listen for specified OS signals.
func (s *Shutdown) Notify(signals ...os.Signal) {
	if s.signalCh == nil {
		s.signalCh = make(chan os.Signal, 1)
	}
	signal.Notify(s.signalCh, signals...)

	go func() {
		sig := <-s.signalCh
		s.Send("os.Signal", sig.String())
		signal.Stop(s.signalCh)
		close(s.signalCh)
	}()
}

// Context returns the context associated with the Shutdown instance.
func (s *Shutdown) Context() context.Context {
	return s.ctx
}

// CancelFunc returns the cancel function for the Shutdown context.
func (s *Shutdown) CancelFunc() context.CancelFunc {
	return s.cancel
}

// Close stops the signal channel and cancels the context.
func (s *Shutdown) Close() {
	if s.signalCh != nil {
		signal.Stop(s.signalCh)
		close(s.signalCh)
	}
	s.cancel()
}

// Wait blocks until the context is done and returns the last message.
func (s *Shutdown) Wait() string {
	<-s.Done()
	return s.Message()
}

func (s *Shutdown) WaitAndDone(wg *sync.WaitGroup, cleanup func()) {
	defer wg.Done()
	<-s.Done()
	if cleanup != nil {
		cleanup()
	}
}