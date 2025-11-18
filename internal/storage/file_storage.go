package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/example/2025-11-18/internal/domain"
)

var ErrBatchNotFound = errors.New("batch not found")

type FileStorage struct {
	mu      sync.RWMutex
	batches map[int]*domain.Batch
	dataDir string
	nextID  int
}

func NewFileStorage(dataDir string) (*FileStorage, error) {
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, fmt.Errorf("создание каталога %s: %w", dataDir, err)
	}

	fs := &FileStorage{
		batches: make(map[int]*domain.Batch),
		dataDir: dataDir,
		nextID:  1,
	}
	if err := fs.loadFromDisk(); err != nil {
		return nil, err
	}
	return fs, nil
}

func (s *FileStorage) loadFromDisk() error {
	entries, err := os.ReadDir(s.dataDir)
	if err != nil {
		return fmt.Errorf("чтение каталога %s: %w", s.dataDir, err)
	}

	var maxID int
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "batch_") || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		id, err := parseBatchID(entry.Name())
		if err != nil {
			continue
		}
		path := filepath.Join(s.dataDir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("чтение файла %s: %w", path, err)
		}
		var batch domain.Batch
		if err := json.Unmarshal(data, &batch); err != nil {
			return fmt.Errorf("разбор файла %s: %w", path, err)
		}
		s.batches[id] = batch.Clone()
		if id > maxID {
			maxID = id
		}
	}
	s.nextID = max(1, maxID+1)
	return nil
}

func parseBatchID(name string) (int, error) {
	var id int
	_, err := fmt.Sscanf(name, "batch_%d.json", &id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (s *FileStorage) CreateBatch(urls []string) (*domain.Batch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	id := s.nextID
	s.nextID++

	batch := &domain.Batch{
		ID:      id,
		URLs:    append([]string(nil), urls...),
		Results: make([]domain.LinkCheckResult, 0, len(urls)),
		Done:    false,
		Created: now,
	}
	s.batches[id] = batch.Clone()
	if err := s.persist(batch); err != nil {
		return nil, err
	}
	return batch.Clone(), nil
}

func (s *FileStorage) UpdateBatchResults(batchID int, results []domain.LinkCheckResult, done bool) (*domain.Batch, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	batch, ok := s.batches[batchID]
	if !ok {
		return nil, ErrBatchNotFound
	}
	batch.Results = append([]domain.LinkCheckResult(nil), results...)
	batch.Done = done
	if err := s.persist(batch); err != nil {
		return nil, err
	}
	return batch.Clone(), nil
}

func (s *FileStorage) GetBatch(id int) (*domain.Batch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	batch, ok := s.batches[id]
	if !ok {
		return nil, ErrBatchNotFound
	}
	return batch.Clone(), nil
}

func (s *FileStorage) GetBatches(ids []int) ([]*domain.Batch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*domain.Batch, 0, len(ids))
	for _, id := range ids {
		batch, ok := s.batches[id]
		if !ok {
			return nil, fmt.Errorf("batch %d: %w", id, ErrBatchNotFound)
		}
		results = append(results, batch.Clone())
	}
	return results, nil
}

func (s *FileStorage) persist(batch *domain.Batch) error {
	path := filepath.Join(s.dataDir, fmt.Sprintf("batch_%d.json", batch.ID))
	data, err := json.MarshalIndent(batch, "", "  ")
	if err != nil {
		return fmt.Errorf("сериализация batch %d: %w", batch.ID, err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, fs.FileMode(0o644)); err != nil {
		return fmt.Errorf("запись %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("обновление %s: %w", path, err)
	}
	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (s *FileStorage) ListBatchIDs() []int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]int, 0, len(s.batches))
	for id := range s.batches {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids
}
