package main

import (
	"log"

	"github.com/codeinium/parcer-golang/internal/config"
	"github.com/codeinium/parcer-golang/internal/scraper"
	"github.com/codeinium/parcer-golang/pkg/storage"
)

func main() {
	log.Println("Загрузка конфигурации...")
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	app := scraper.NewScraper(cfg)

	log.Println("Запуск скрейпера...")
	products, err := app.Run()
	if err != nil {
		log.Fatalf("Ошибка в процессе сбора данных: %v", err)
	}

	if len(products) == 0 {
		log.Println("Не найдено ни одного товара. Проверьте настройки категорий в config.yaml.")
		return
	}

	log.Printf("Сбор данных завершен. Всего найдено %d товаров.", len(products))

	if err := storage.WriteProductsToCSV(products, cfg.OutputFile); err != nil {
		log.Fatalf("Ошибка сохранения данных в CSV: %v", err)
	}

	log.Printf("Данные успешно сохранены в файл: %s", cfg.OutputFile)
}
