package domain

import "time"

type LinkCheckResult struct {
	URL      string `json:"url"`
	Status   string `json:"status"` // "available" / "unavailable"
	HTTPCode int    `json:"http_code,omitempty"`
	Error    string `json:"error,omitempty"`
}

type Batch struct {
	ID      int               `json:"id"`
	URLs    []string          `json:"urls"`
	Results []LinkCheckResult `json:"results,omitempty"`
	Done    bool              `json:"done"`
	Created time.Time         `json:"created"`
}

func (b *Batch) Clone() *Batch {
	if b == nil {
		return nil
	}
	copyBatch := *b
	if b.URLs != nil {
		copyBatch.URLs = append([]string(nil), b.URLs...)
	}
	if b.Results != nil {
		copyBatch.Results = append([]LinkCheckResult(nil), b.Results...)
	}
	return &copyBatch
}
