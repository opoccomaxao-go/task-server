package storage

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/opoccomaxao-go/task-server/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ImplementationTest(t *testing.T, storage task.Storage) {
	watcherCalls := int64(0)

	iter, err := storage.Watch()
	require.NoError(t, err)

	data := json.RawMessage(`{"url":"1","data":{"a":1,"b":"2"},"2":2}`)

	toSave := task.Task{
		Data:       data,
		Expiration: time.Now().Add(time.Hour).UTC().Truncate(time.Second),
		Executed:   false,
	}

	go func() {
		_ = iter.Wait(context.Background())
		watcherCalls++
	}()

	err = storage.Create(toSave)
	require.NoError(t, err)

	time.Sleep(time.Second)
	assert.Equal(t, int64(1), watcherCalls)

	tTask, err := storage.FirstToExecute()
	require.NoError(t, err)
	require.NotNil(t, tTask)

	newData := tTask.Data
	toSave.ID = tTask.ID
	toSave.Data = nil
	tTask.Data = nil
	assert.Equal(t, &toSave, tTask)
	assert.JSONEq(t, string(data), string(newData))

	toSave.Executed = true
	toSave.Data = data

	go func() {
		_ = iter.Wait(context.Background())
		watcherCalls++
	}()

	err = storage.Update(toSave)
	require.NoError(t, err)

	time.Sleep(time.Second)
	assert.Equal(t, int64(2), watcherCalls)

	tTask, err = storage.FirstToExecute()
	require.Error(t, err)
	require.Nil(t, tTask)
	assert.True(t, errors.Is(err, task.ErrEmpty), "implementation must return task.ErrEmpty when no tasks")

	require.NoError(t, storage.Close())
}
