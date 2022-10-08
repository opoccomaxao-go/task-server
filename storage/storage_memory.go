package storage

import (
	"context"
	"sync"
	"time"

	"github.com/opoccomaxao-go/task-server/notificator"
	"github.com/opoccomaxao-go/task-server/task"
)

type storageMemory struct {
	mu sync.Mutex

	buffer []*task.Task
	mapped map[task.ID]*task.Task
	ticker notificator.Extended
	noty   *notificator.Storage
}

func NewMemory() task.Storage {
	res := storageMemory{
		buffer: make([]*task.Task, 1000),
		mapped: make(map[task.ID]*task.Task, 1000),
		ticker: notificator.New().NotifyEveryTick(time.Minute),
		noty:   notificator.NewStorage(),
	}

	go res.taskClean()

	return &res
}

func (s *storageMemory) Create(t task.Task) error {
	s.mu.Lock()
	t.ID = task.NewID()
	s.buffer = append(s.buffer, &t)
	s.mapped[t.ID] = &t
	s.mu.Unlock()

	go s.noty.NotifyAll()

	return nil
}

func (s *storageMemory) Update(t task.Task) error {
	s.mu.Lock()
	if ptr, ok := s.mapped[t.ID]; ok {
		*ptr = t
	}
	s.mu.Unlock()

	go s.noty.NotifyAll()

	return nil
}

func (s *storageMemory) FirstToExecute() (*task.Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, t := range s.buffer {
		if t != nil && !t.Executed {
			res := *t

			return &res, nil
		}
	}

	return nil, task.ErrEmpty
}

func (s *storageMemory) Watch() (task.Notificator, error) {
	noty := notificator.NewWithCapacity(10)
	s.noty.Add(noty)

	return noty, nil
}

func (s *storageMemory) Close() error {
	s.mu.Lock()
	_ = s.ticker.Close(context.Background())
	s.buffer = nil
	s.mapped = nil
	s.mu.Unlock()

	return nil
}

func (s *storageMemory) taskClean() {
	var err error
	for err == nil {
		now := time.Now()

		s.mu.Lock()

		for i, task := range s.buffer {
			if task != nil && task.Expiration.Before(now) {
				s.buffer[i] = nil
				delete(s.mapped, task.ID)
			}
		}

		for i, task := range s.buffer {
			if task != nil {
				s.buffer = s.buffer[i:]

				break
			}
		}

		s.mu.Unlock()

		err = s.ticker.Wait(context.Background())
	}
}
