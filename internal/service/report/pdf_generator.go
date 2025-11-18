package report

import (
	"bytes"
	"fmt"

	"github.com/jung-kurt/gofpdf"

	"github.com/example/2025-11-18/internal/domain"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Build(batches []*domain.Batch) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 20, 15)
	pdf.SetAutoPageBreak(true, 15)

	for i, batch := range batches {
		pdf.AddPage()
		pdf.SetFont("Arial", "B", 16)
		pdf.Cell(0, 10, fmt.Sprintf("Batch #%d", batch.ID))
		pdf.Ln(12)

		pdf.SetFont("Arial", "", 12)
		pdf.Cell(0, 8, fmt.Sprintf("Создан: %s", batch.Created.Format("2006-01-02 15:04:05 MST")))
		pdf.Ln(8)
		status := "Завершён"
		if !batch.Done {
			status = "В обработке"
		}
		pdf.Cell(0, 8, fmt.Sprintf("Статус: %s", status))
		pdf.Ln(12)

		pdf.SetFont("Arial", "B", 11)
		headers := []string{"URL", "Статус", "HTTP", "Комментарий"}
		widths := []float64{90, 30, 20, 40}
		for idx, header := range headers {
			pdf.CellFormat(widths[idx], 8, header, "1", 0, "", false, 0, "")
		}
		pdf.Ln(-1)

		pdf.SetFont("Arial", "", 10)
		if len(batch.Results) == 0 {
			pdf.CellFormat(0, 8, "Результаты отсутствуют", "1", 0, "", false, 0, "")
			pdf.Ln(-1)
		} else {
			for _, result := range batch.Results {
				pdf.CellFormat(widths[0], 8, truncate(result.URL, 70), "1", 0, "", false, 0, "")
				pdf.CellFormat(widths[1], 8, result.Status, "1", 0, "", false, 0, "")
				httpCode := ""
				if result.HTTPCode != 0 {
					httpCode = fmt.Sprintf("%d", result.HTTPCode)
				}
				pdf.CellFormat(widths[2], 8, httpCode, "1", 0, "", false, 0, "")
				pdf.CellFormat(widths[3], 8, truncate(result.Error, 40), "1", 0, "", false, 0, "")
				pdf.Ln(-1)
			}
		}

		if i < len(batches)-1 {
			pdf.Ln(5)
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	if limit <= 3 {
		return s[:limit]
	}
	return s[:limit-3] + "..."
}
