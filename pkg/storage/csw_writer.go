package storage

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/codeinium/parcer-golang/internal/models"
)

func WriteProductsToCSV(products []models.Product, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("не удалось создать файл: %w", err)
	}
	defer file.Close()
	if _, err := file.WriteString("\xEF\xBB\xBF"); err != nil {
		return err
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"Наименование", "Цена", "Ссылка", "Категория"}
	if err := writer.Write(headers); err != nil {
		return err
	}

	for _, product := range products {
		row := []string{product.Name, product.Price, product.URL, product.Category}
		if err := writer.Write(row); err != nil {
			continue
		}
	}

	return writer.Error()
}
