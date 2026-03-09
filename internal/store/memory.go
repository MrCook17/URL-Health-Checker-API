package store

import (
	"sync"

	"healthcheck-api/internal/model"
)

type MemoryStore struct {
	mu   sync.RWMutex
	jobs map[string]model.CheckJob
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		jobs: make(map[string]model.CheckJob),
	}
}

func (s *MemoryStore) Create(job model.CheckJob) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs[job.ID] = job
}

func (s *MemoryStore) Get(id string) (model.CheckJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, ok := s.jobs[id]
	return job, ok
}