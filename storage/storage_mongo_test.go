package storage

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/opoccomaxao-go/task-server/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestStorageMongo(t *testing.T) {
	t.Parallel()

	storage, err := NewMongo(StorageMongoConfig{
		ConnectURL: "mongodb://localhost:27017",
		DBName:     "test",
	})
	require.NoError(t, err)

	ImplementationTest(t, storage)
}

func TestMongoTaskToTask(t *testing.T) {
	t.Parallel()

	tTime := time.Now().UTC().Truncate(time.Second)
	tMongoTask := mongoTask{
		ID:         1,
		Expiration: tTime,
		Executed:   true,
		Data: primitive.D{
			{Key: "a", Value: "1"},
			{Key: "b", Value: 2},
			{Key: "c", Value: primitive.D{
				{Key: "a", Value: 1},
				{Key: "b", Value: "2"},
			}},
		},
	}

	tTask := tMongoTask.Task()
	assert.JSONEq(t, `{"a":"1","b":2,"c":{"a":1,"b":"2"}}`, string(tTask.Data))

	tTask.Data = nil
	assert.Equal(t, &task.Task{
		ID:         1,
		Expiration: tTime,
		Executed:   true,
	}, tTask)
}

func TestTaskToMongoTask(t *testing.T) {
	t.Parallel()

	tTime := time.Now().UTC().Truncate(time.Second)
	tTask := task.Task{
		ID:         1,
		Expiration: tTime,
		Executed:   true,
		Data:       json.RawMessage(`{"a":"1","b":2,"c":{"a":1,"b":"2"}}`),
	}

	tMongoTask := taskToMongoTask(tTask)

	assert.Equal(t, mongoTask{
		ID:         1,
		Expiration: tTime,
		Executed:   true,
		Data: map[string]interface{}{
			"a": "1",
			"b": 2.,
			"c": map[string]interface{}{
				"a": 1.,
				"b": "2",
			},
		},
	}, tMongoTask)
}
