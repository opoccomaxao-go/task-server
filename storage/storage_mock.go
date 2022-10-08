package storage

import (
	"github.com/opoccomaxao-go/task-server/notificator"
	"github.com/opoccomaxao-go/task-server/task"
)

type StorageMock struct {
	First   chan task.Task
	Created chan task.Task
	Updated chan task.Task
}

var _ task.Storage = (*StorageMock)(nil)

func NewMock() *StorageMock {
	return &StorageMock{
		First:   make(chan task.Task, 10000),
		Created: make(chan task.Task, 10000),
		Updated: make(chan task.Task, 10000),
	}
}

func (m *StorageMock) FillFirst(tasks []task.Task) {
	for _, t := range tasks {
		m.First <- t
	}
}

func (m *StorageMock) Create(t task.Task) error {
	m.Created <- t

	return nil
}

func (m *StorageMock) Update(t task.Task) error {
	m.Updated <- t

	return nil
}

func (m *StorageMock) FirstToExecute() (*task.Task, error) {
	select {
	case res, ok := <-m.First:
		if ok {
			return &res, nil
		}

		return nil, task.ErrClosed
	default:
		return nil, task.ErrEmpty
	}
}

func (m *StorageMock) Watch() (task.Notificator, error) {
	return notificator.New(), nil
}

func (m *StorageMock) Close() error {
	close(m.Created)
	close(m.First)
	close(m.Updated)

	return nil
}
