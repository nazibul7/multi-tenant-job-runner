package store

import "sync"

type IdempotencyRecord struct {
	RequestHash string
	JobID       string
}

type IdempotencyStore struct {
	mu      sync.RWMutex
	records map[string]IdempotencyRecord
}

func NewIdempotencyStore() *IdempotencyStore {
	return &IdempotencyStore{
		records: make(map[string]IdempotencyRecord),
	}
}

func (s *IdempotencyStore) Get(key string) (IdempotencyRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.records[key]
	return record, ok
}

func (s *IdempotencyStore) Save(key string, record IdempotencyRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records[key] = record
}
