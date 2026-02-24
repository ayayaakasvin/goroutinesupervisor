package goroutinesupervisor

import (
	"context"
	"time"
)

type Task struct {
	Name string
	Run  func(ctx context.Context) error
}

type TaskRunner func(ctx context.Context) error

type Interrupter func(error)

type EventType int

const (
	EventTaskStarted EventType = iota
	EventTaskFinished
	EventTaskFailed
)

type Event struct {
	Type    EventType
	Task    string
	Err     error
	Panic   bool
	Started time.Time
	Ended   time.Time
}

type EventHandler func(Event)
