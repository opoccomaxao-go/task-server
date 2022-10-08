package task

import (
	"encoding/json"
	"time"
)

type ID uint64

func NewID() ID {
	return ID(time.Now().UnixNano())
}

type Task struct {
	ID         ID
	Expiration time.Time
	Executed   bool
	Data       json.RawMessage
}
