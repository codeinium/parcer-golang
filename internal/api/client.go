package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/codeinium/parcer-golang/internal/config"
	"github.com/codeinium/parcer-golang/internal/models"
)

type Client struct {
	httpClient *http.Client
}

func NewClient(cfg *config.Config) (*Client, error) {
	transport := &http.Transport{}

	if cfg.Proxy.Enabled {
		proxyURL, err := url.Parse(cfg.Proxy.URL)
		if err != nil {
			return nil, fmt.Errorf("неверный URL прокси: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	return &Client{
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}, nil
}

func (c *Client) FetchShowcase(lat, lon float64) (*models.ShowcaseResponse, error) {
	fullURL := fmt.Sprintf("%s?lat=%f&lon=%f", showcaseURL, lat, lon)

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("User-Agent", "okhttp/4.9.0")
	req.Header.Set("x-app-version", "5.22.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка выполнения запроса к API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API вернул неожиданный статус: %s", resp.Status)
	}

	var showcase models.ShowcaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&showcase); err != nil {
		return nil, fmt.Errorf("ошибка декодирования JSON: %w", err)
	}

	return &showcase, nil
}
