package worker_pool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const defaultPolzaURL = "https://polza.ai/api/v1/chat/completions"

type StaticPlanner struct{}

func (StaticPlanner) GeneratePlan(_ context.Context, task Task) (string, error) {
	title := strings.TrimSpace(task.Title)
	if title == "" {
		title = "Unnamed task"
	}

	return fmt.Sprintf(
		"1. Уточнить требования и ограничения по задаче \"%s\".\n2. Разбить реализацию на backend, data и API слои.\n3. Описать изменения в схеме данных и контрактах.\n4. Реализовать код и покрыть критический путь тестами.\n5. Проверить сценарии ошибки, логирование и выкладку.",
		title,
	), nil
}

type PolzaPlanner struct {
	client   *http.Client
	endpoint string
	apiKey   string
	model    string
}

func NewPlanGeneratorFromEnv() PlanGenerator {
	apiKey := strings.TrimSpace(os.Getenv("POLZA_API_KEY"))
	if apiKey == "" {
		return StaticPlanner{}
	}

	endpoint := strings.TrimSpace(os.Getenv("POLZA_API_URL"))
	if endpoint == "" {
		endpoint = defaultPolzaURL
	}

	model := strings.TrimSpace(os.Getenv("POLZA_MODEL"))
	if model == "" {
		model = "openai/gpt-5.4-mini"
	}

	return &PolzaPlanner{
		client: &http.Client{
			Timeout: 40 * time.Second,
		},
		endpoint: endpoint,
		apiKey:   apiKey,
		model:    model,
	}
}

func (p *PolzaPlanner) GeneratePlan(ctx context.Context, task Task) (string, error) {
	payload := chatCompletionRequest{
		Model: p.model,
		Messages: []chatMessage{
			{
				Role:    "system",
				Content: "Ты senior software engineer. На вход приходит короткое название backend-задачи. Верни краткий и конкретный план реализации на русском языке в 4-7 шагах без вводных слов.",
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Сформируй план реализации для задачи: %s", strings.TrimSpace(task.Title)),
			},
		},
		Temperature: 0.2,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal llm request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("build llm request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("call llm: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("read llm response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("llm status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed chatCompletionResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("decode llm response: %w", err)
	}

	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("llm returned no choices")
	}

	content := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if content == "" {
		return "", fmt.Errorf("llm returned empty content")
	}

	return content, nil
}

type chatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}
