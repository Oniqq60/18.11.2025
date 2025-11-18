package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/example/2025-11-18/internal/service/report"
	"github.com/example/2025-11-18/internal/storage"
)

type ReportHandler struct {
	storage   *storage.FileStorage
	generator *report.Generator
}

func NewReportHandler(storage *storage.FileStorage, generator *report.Generator) *ReportHandler {
	return &ReportHandler{
		storage:   storage,
		generator: generator,
	}
}

type reportRequest struct {
	LinksNum []int `json:"links_num"`
}

func (h *ReportHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	var req reportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "некорректное тело запроса", http.StatusBadRequest)
		return
	}
	if len(req.LinksNum) == 0 {
		http.Error(w, "укажите хотя бы один batch_id", http.StatusBadRequest)
		return
	}

	batches, err := h.storage.GetBatches(req.LinksNum)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	pdfBytes, err := h.generator.Build(batches)
	if err != nil {
		http.Error(w, "не удалось собрать отчёт", http.StatusInternalServerError)
		return
	}

	filename := buildReportFilename(req.LinksNum)
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(pdfBytes)
}

func buildReportFilename(ids []int) string {
	values := make([]string, len(ids))
	for i, id := range ids {
		values[i] = strconv.Itoa(id)
	}
	return "report_" + strings.Join(values, "_") + ".pdf"
}
