package notificator

import (
	"context"
	"strings"
	"sync"

	"github.com/opoccomaxao-go/task-server/task"
	"github.com/pkg/errors"
)

type Storage struct {
	data []task.Notificator
	mu   sync.Mutex
}

func NewStorage() *Storage {
	return &Storage{}
}

func (s *Storage) Add(noty task.Notificator) {
	s.mu.Lock()
	s.data = append(s.data, noty)
	s.mu.Unlock()
}

func (s *Storage) Close() error {
	var errs []error

	s.mu.Lock()
	for _, n := range s.data {
		err := n.Close(context.Background())
		if err != nil {
			errs = append(errs, err)
		}
	}
	s.mu.Unlock()

	if len(errs) > 0 {
		messages := make([]string, len(errs))
		for i, err := range errs {
			messages[i] = err.Error()
		}

		return errors.New(strings.Join(messages, "\n"))
	}

	return nil
}

// NotifyAll apply Notify to each element and filters closed.
func (s *Storage) NotifyAll() {
	s.mu.Lock()
	original := s.data
	s.data = s.data[0:0]

	for _, it := range original {
		if err := it.Notify(); err == nil {
			s.data = append(s.data, it)
		}
	}
	s.mu.Unlock()
}
