package scraper

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/codeinium/parcer-golang/internal/config"
	"github.com/codeinium/parcer-golang/internal/models"
)

const baseURL = "https://samokat.ru"

type Scraper struct {
	config *config.Config
}

func NewScraper(cfg *config.Config) *Scraper {
	return &Scraper{config: cfg}
}

func (s *Scraper) Run() ([]models.Product, error) {
	opts := []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.DisableGPU,
	}
	if s.config.Headless {
		opts = append(opts, chromedp.Headless)
	}
	if s.config.Proxy.Enabled {
		log.Printf("Используется прокси: %s", s.config.Proxy.Server)
		opts = append(opts, chromedp.ProxyServer(s.config.Proxy.Server))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.config.TimeoutSeconds)*time.Second)
	defer cancel()

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	log.Println("Запуск браузера и автоматический выбор адреса...")
	if err := s.selectAddress(taskCtx); err != nil {
		var b []byte
		if errShot := chromedp.Run(taskCtx, chromedp.CaptureScreenshot(&b)); errShot == nil {
			log.Printf("Скриншот ошибки сохранен в error_screenshot.png")
			_ = os.WriteFile("error_screenshot.png", b, 0644)
		}
		return nil, fmt.Errorf("не удалось автоматически выбрать адрес: %w", err)
	}

	var allProducts []models.Product
	for _, categoryName := range s.config.TargetCategories {
		log.Printf("--- Обработка категории: %s ---", categoryName)
		if err := chromedp.Run(taskCtx, chromedp.Navigate(baseURL)); err != nil {
			log.Printf("Не удалось вернуться на главную страницу: %v", err)
			continue
		}
		products, err := s.scrapeCategory(taskCtx, categoryName)
		if err != nil {
			log.Printf("ПРЕДУПРЕЖДЕНИЕ: Не удалось обработать категорию '%s': %v", categoryName, err)
			continue
		}
		allProducts = append(allProducts, products...)
	}

	return allProducts, nil
}

func (s *Scraper) selectAddress(ctx context.Context) error {
	addressInputSelector := `//input[@placeholder="Введите адрес доставки"]`
	suggestionListSelector := `//div[contains(@class, "AddressSuggestions_root")]/div[1]`

	return chromedp.Run(ctx,
		chromedp.Navigate(baseURL),

		chromedp.ActionFunc(func(c context.Context) error {
			log.Println("1/4: Ожидание поля для ввода адреса...")
			return nil
		}),
		chromedp.WaitVisible(addressInputSelector, chromedp.BySearch),

		chromedp.ActionFunc(func(c context.Context) error {
			log.Println("2/4: Ввод адреса и ожидание подсказок...")
			return nil
		}),
		chromedp.SendKeys(addressInputSelector, s.config.TargetAddress, chromedp.BySearch),
		chromedp.WaitVisible(suggestionListSelector, chromedp.BySearch),

		chromedp.ActionFunc(func(c context.Context) error {
			log.Println("3/4: Клик по первой подсказке...")
			return nil
		}),
		chromedp.Click(suggestionListSelector, chromedp.BySearch),

		chromedp.ActionFunc(func(c context.Context) error {
			log.Println("4/4: Ожидание обновления страницы с адресом...")
			return nil
		}),
		chromedp.WaitVisible(`[data-qa-id="address_button_description_text"]`, chromedp.ByQuery),

		chromedp.ActionFunc(func(c context.Context) error {
			log.Println("Адрес доставки успешно установлен!")
			return nil
		}),
	)
}

func (s *Scraper) scrapeCategory(ctx context.Context, categoryName string) ([]models.Product, error) {
	var categoryURL string
	selector := fmt.Sprintf(`//a[.//h3[text()="%s"]]`, categoryName)
	findLinkTask := chromedp.Tasks{
		chromedp.WaitVisible(selector, chromedp.BySearch),
		chromedp.AttributeValue(selector, "href", &categoryURL, nil, chromedp.BySearch),
	}
	if err := chromedp.Run(ctx, findLinkTask); err != nil {
		return nil, fmt.Errorf("ссылка на категорию '%s' не найдена: %w", categoryName, err)
	}
	log.Printf("Найдена ссылка: %s. Переход...", categoryURL)
	fullCategoryURL := baseURL + categoryURL
	var nodes []*cdp.Node
	err := chromedp.Run(ctx,
		chromedp.Navigate(fullCategoryURL),
		chromedp.WaitVisible(`//article`, chromedp.BySearch),
		chromedp.Nodes(`//article`, &nodes, chromedp.BySearch),
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось загрузить товары на странице категории: %w", err)
	}
	log.Printf("Найдено %d карточек товаров. Идет сбор данных...", len(nodes))
	var products []models.Product
	for _, node := range nodes {
		var name, price, href string
		err := chromedp.Run(ctx,
			chromedp.Text(`h3[itemprop="name"]`, &name, chromedp.FromNode(node), chromedp.ByQuery),
			chromedp.Text(`div[itemprop="price"]`, &price, chromedp.FromNode(node), chromedp.ByQuery),
			chromedp.AttributeValue(`a[itemprop="url"]`, "href", &href, nil, chromedp.FromNode(node), chromedp.ByQuery),
		)
		if err != nil {
			log.Printf("Не удалось распарсить карточку: %v", err)
			continue
		}
		products = append(products, models.Product{
			Category: categoryName,
			Name:     strings.TrimSpace(name),
			Price:    strings.TrimSpace(strings.ReplaceAll(price, "₽", "")),
			URL:      baseURL + href,
		})
	}
	return products, nil
}
