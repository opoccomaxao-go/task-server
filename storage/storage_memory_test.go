package storage

import (
	"testing"
)

func TestStorageMemory(t *testing.T) {
	t.Parallel()

	ImplementationTest(t, NewMemory())
}
