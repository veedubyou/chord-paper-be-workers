package dummy

import (
	"chord-paper-be-workers/src/application/cloud_storage/entity"
	"context"
	"sync"
)

var _ entity.FileStore = &FileStore{}

func NewDummyFileStore() *FileStore {
	return &FileStore{
		Unavailable: false,
		State:       make(map[string][]byte),
	}
}

type FileStore struct {
	Unavailable bool
	State       map[string][]byte
	mutex       sync.RWMutex
}

func (t *FileStore) GetFile(_ context.Context, url string) ([]byte, error) {
	if t.Unavailable {
		return nil, NetworkFailure
	}

	t.mutex.RLock()
	defer t.mutex.RUnlock()

	content, ok := t.State[url]
	if !ok {
		return nil, NotFound
	}

	return content, nil
}

func (t *FileStore) WriteFile(_ context.Context, url string, fileContent []byte) error {
	if t.Unavailable {
		return NetworkFailure
	}

	t.mutex.Lock()
	defer t.mutex.Unlock()
	t.State[url] = append([]byte{}, fileContent...)

	return nil
}
