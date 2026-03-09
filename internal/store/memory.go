package store

import (
	"sync"

	"healthcheck-api/internal/model"
)

// MemoryStore keeps jobs in RAM only.
// It is suitable for development and early project stages.
type MemoryStore struct {
	mu   sync.RWMutex
	jobs map[string]model.CheckJob
}

// NewMemoryStore creates an empty in-memory job store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		jobs: make(map[string]model.CheckJob),
	}
}

// Create stores a new check job by ID.
func (s *MemoryStore) Create(job model.CheckJob) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs[job.ID] = job
}

// Get returns a job by ID and whether it exists.
func (s *MemoryStore) Get(id string) (model.CheckJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, ok := s.jobs[id]
	return job, ok
}