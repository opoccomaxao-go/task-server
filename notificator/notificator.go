package notificator

import (
	"context"
	"time"

	"github.com/opoccomaxao-go/task-server/task"
	"github.com/pkg/errors"
)

type Extended interface {
	task.Notificator

	NotifyEveryTick(time.Duration) Extended
}

type notificator struct {
	channel chan struct{}
	closed  bool
}

func (n *notificator) Wait(ctx context.Context) error {
	if n.closed {
		return task.ErrClosed
	}

	select {
	case <-ctx.Done():
		return errors.WithStack(ctx.Err())
	case _, ok := <-n.channel:
		if ok {
			return nil
		}

		return task.ErrClosed
	}
}

func (n *notificator) Close(context.Context) error {
	if n.closed {
		return nil
	}

	close(n.channel)

	return nil
}

func (n *notificator) Notify() error {
	if n.closed {
		return task.ErrClosed
	}

	select {
	case n.channel <- struct{}{}:
	default:
	}

	return nil
}

func (n *notificator) notifyEveryTick(d time.Duration) {
	ticker := time.NewTicker(d)
	defer ticker.Stop()

	for range ticker.C {
		err := n.Notify()
		if err != nil {
			return
		}
	}
}

func (n *notificator) NotifyEveryTick(duration time.Duration) Extended {
	go n.notifyEveryTick(duration)

	return n
}

func New() Extended {
	return NewWithCapacity(100)
}

func NewWithCapacity(capacity int) Extended {
	return &notificator{
		channel: make(chan struct{}, capacity),
		closed:  false,
	}
}
