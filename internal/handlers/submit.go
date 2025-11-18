package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/example/2025-11-18/internal/domain"
	"github.com/example/2025-11-18/internal/service/checker"
	"github.com/example/2025-11-18/internal/storage"
)

type SubmitHandler struct {
	storage *storage.FileStorage
	checker *checker.Service
}

func NewSubmitHandler(storage *storage.FileStorage, checker *checker.Service) *SubmitHandler {
	return &SubmitHandler{
		storage: storage,
		checker: checker,
	}
}

type submitRequest struct {
	Links []string `json:"links"`
}

type submitResponse struct {
	Links    map[string]string `json:"links"`
	LinksNum int               `json:"links_num"`
}

func (h *SubmitHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req submitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "некорректное тело запроса", http.StatusBadRequest)
		return
	}

	if len(req.Links) == 0 {
		http.Error(w, "список ссылок пуст", http.StatusBadRequest)
		return
	}
	if len(req.Links) > domain.MaxURLsPerBatch {
		http.Error(w, "слишком много ссылок", http.StatusBadRequest)
		return
	}

	type preparedLink struct {
		original   string
		normalized string
	}
	validated := make([]preparedLink, 0, len(req.Links))
	normURLs := make([]string, 0, len(req.Links))
	for _, raw := range req.Links {
		trimmed := strings.TrimSpace(raw)
		if trimmed == "" {
			http.Error(w, "обнаружена пустая ссылка", http.StatusBadRequest)
			return
		}
		parsed, err := domain.ValidateURL(trimmed)
		if err != nil {
			http.Error(w, "найден некорректный URL", http.StatusBadRequest)
			return
		}
		validated = append(validated, preparedLink{
			original:   trimmed,
			normalized: parsed.String(),
		})
		normURLs = append(normURLs, parsed.String())
	}

	batch, err := h.storage.CreateBatch(normURLs)
	if err != nil {
		http.Error(w, "не удалось создать batch", http.StatusInternalServerError)
		return
	}

	results, err := h.checker.ProcessBatch(r.Context(), batch.ID, batch.URLs)
	if err != nil {
		http.Error(w, "не удалось проверить ссылки", http.StatusInternalServerError)
		return
	}

	responseLinks := make(map[string]string, len(validated))
	for i, link := range validated {
		status := "not available"
		if i < len(results) && results[i].Status == domain.StatusAvailable {
			status = "available"
		}
		responseLinks[link.original] = status
	}

	writeJSON(w, http.StatusOK, submitResponse{
		Links:    responseLinks,
		LinksNum: batch.ID,
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil && !errors.Is(err, context.Canceled) {
		http.Error(w, "ошибка сериализации ответа", http.StatusInternalServerError)
	}
}
