package checker

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/example/2025-11-18/internal/domain"
	"github.com/example/2025-11-18/internal/storage"
)

type Service struct {
	storage *storage.FileStorage
	client  *http.Client
}

func New(storage *storage.FileStorage, timeout time.Duration) *Service {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &Service{
		storage: storage,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (s *Service) ProcessBatch(ctx context.Context, batchID int, urls []string) ([]domain.LinkCheckResult, error) {
	results := make([]domain.LinkCheckResult, 0, len(urls))
	for _, raw := range urls {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}
		result := s.checkURL(ctx, raw)
		results = append(results, result)
		if _, err := s.storage.UpdateBatchResults(batchID, results, false); err != nil {
			return results, err
		}
	}
	if _, err := s.storage.UpdateBatchResults(batchID, results, true); err != nil {
		return results, err
	}
	return results, nil
}

func (s *Service) checkURL(ctx context.Context, raw string) domain.LinkCheckResult {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return domain.LinkCheckResult{
			URL:    raw,
			Status: domain.StatusUnavailable,
			Error:  err.Error(),
		}
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return domain.LinkCheckResult{
			URL:    raw,
			Status: domain.StatusUnavailable,
			Error:  err.Error(),
		}
	}
	defer resp.Body.Close()

	status := domain.StatusUnavailable
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		status = domain.StatusAvailable
	}
	result := domain.LinkCheckResult{
		URL:      raw,
		Status:   status,
		HTTPCode: resp.StatusCode,
	}
	if status == domain.StatusUnavailable {
		result.Error = fmt.Sprintf("unexpected status %d", resp.StatusCode)
	}
	return result
}
