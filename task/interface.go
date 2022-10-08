package task

import (
	"context"
)

// Notificator common implementation located at:
//
//	"github.com/opoccomaxao-go/task-server/notificator".
type Notificator interface {
	Wait(context.Context) error
	Close(context.Context) error
	Notify() error
}

// Storage common implementation located at:
//
//	"github.com/opoccomaxao-go/task-server/storage".
type Storage interface {
	Create(Task) error
	Update(Task) error
	// if no documents ErrEmpty should be returned.
	FirstToExecute() (*Task, error)
	Watch() (Notificator, error)
	Close() error
}
