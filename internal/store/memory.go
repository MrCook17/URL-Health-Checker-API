package store

import (
	"sort"
	"sync"

	"healthcheck-api/internal/model"
)

// MemoryStore keeps check jobs in memory only.
// It is simple and suitable for the current project stage.
type MemoryStore struct {
	mu   sync.RWMutex
	jobs map[string]model.CheckJob
}

// NewMemoryStore creates an empty in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		jobs: make(map[string]model.CheckJob),
	}
}

// Create stores a newly created job.
func (s *MemoryStore) Create(job model.CheckJob) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs[job.ID] = job
}

// Update replaces an existing job entry with its latest state.
func (s *MemoryStore) Update(job model.CheckJob) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs[job.ID] = job
}

// Get returns one job by ID and whether it exists.
func (s *MemoryStore) Get(id string) (model.CheckJob, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	job, ok := s.jobs[id]
	return job, ok
}

// List returns all stored jobs in sorted ID order so output is predictable.
func (s *MemoryStore) List() []model.CheckJob {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.jobs))
	for id := range s.jobs {
		ids = append(ids, id)
	}

	sort.Strings(ids)

	jobs := make([]model.CheckJob, 0, len(ids))
	for _, id := range ids {
		jobs = append(jobs, s.jobs[id])
	}

	return jobs
}