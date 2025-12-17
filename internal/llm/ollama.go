// Package llm provides integration with local LLMs for text correction.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	DefaultOllamaURL = "http://localhost:11434"
	DefaultModel     = "qwen2.5:0.5b"
	DefaultTimeout   = 10 * time.Second
)

// Client представляет клиент для работы с Ollama.
type Client struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// Config конфигурация LLM клиента.
type Config struct {
	Enabled bool
	URL     string
	Model   string
	Timeout time.Duration
}

// DefaultConfig возвращает конфигурацию по умолчанию.
func DefaultConfig() Config {
	return Config{
		Enabled: false,
		URL:     DefaultOllamaURL,
		Model:   DefaultModel,
		Timeout: DefaultTimeout,
	}
}

// New создаёт новый LLM клиент.
func New(cfg Config) *Client {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = DefaultTimeout
	}

	url := cfg.URL
	if url == "" {
		url = DefaultOllamaURL
	}

	model := cfg.Model
	if model == "" {
		model = DefaultModel
	}

	return &Client{
		baseURL: url,
		model:   model,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// generateRequest запрос к Ollama API.
type generateRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Options struct {
		Temperature float64 `json:"temperature"`
		NumPredict  int     `json:"num_predict"`
	} `json:"options"`
}

// generateResponse ответ от Ollama API.
type generateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

// CorrectText исправляет текст с помощью LLM.
func (c *Client) CorrectText(ctx context.Context, text string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return text, nil
	}

	prompt := fmt.Sprintf(`Исправь ошибки распознавания речи в тексте. Верни ТОЛЬКО исправленный текст без пояснений:

%s`, text)

	req := generateRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: false,
	}
	req.Options.Temperature = 0.1 // Низкая температура для стабильного результата
	req.Options.NumPredict = 500  // Ограничение длины ответа

	body, err := json.Marshal(req)
	if err != nil {
		return text, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return text, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	log.Printf("LLM: отправка запроса на исправление (%d символов)", len(text))
	start := time.Now()

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return text, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return text, fmt.Errorf("ollama error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return text, fmt.Errorf("decode response: %w", err)
	}

	if result.Error != "" {
		return text, fmt.Errorf("ollama: %s", result.Error)
	}

	corrected := strings.TrimSpace(result.Response)
	log.Printf("LLM: исправлено за %v: %q -> %q", time.Since(start).Round(time.Millisecond), text, corrected)

	return corrected, nil
}

// IsAvailable проверяет доступность Ollama.
func (c *Client) IsAvailable(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return false
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// ListModels возвращает список доступных моделей.
func (c *Client) ListModels(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	models := make([]string, len(result.Models))
	for i, m := range result.Models {
		models[i] = m.Name
	}

	return models, nil
}

// Model возвращает текущую модель.
func (c *Client) Model() string {
	return c.model
}

// SetModel устанавливает модель.
func (c *Client) SetModel(model string) {
	c.model = model
}
