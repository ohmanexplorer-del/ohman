package ai

import (
	"context"
	"fmt"
	"log"
	"strings"

	"ohman/src/core/prompts"

	"github.com/sashabaranov/go-openai"
)

type Service struct {
	client *openai.Client
	config AIConfig
}

type AIConfig struct {
	APIKey      string
	BaseURL     string
	Model       string
	MaxTokens   int
	Temperature float64
	TopP        float64
	Stream      bool
}

func NewService(config AIConfig) (*Service, error) {
	if config.BaseURL == "" {
		return nil, fmt.Errorf("ai.base_url is required")
	}
	if config.Model == "" {
		config.Model = "combo"
	}
	if config.MaxTokens == 0 {
		config.MaxTokens = 300
	}
	if config.APIKey == "" {
		config.APIKey = "genai-gtw"
	}

	clientConfig := openai.DefaultConfig(config.APIKey)
	clientConfig.BaseURL = openAIBaseURL(config.BaseURL)
	log.Printf("AI service initialized with OpenAI SDK base URL: %s, model: %s", clientConfig.BaseURL, config.Model)

	return &Service{
		client: openai.NewClientWithConfig(clientConfig),
		config: config,
	}, nil
}

func (s *Service) GenerateContent(ctx context.Context, prompt string) (string, error) {
	systemPrompt, err := prompts.SystemPrompt()
	if err != nil {
		log.Printf("Warning: failed to load system prompt: %v", err)
		systemPrompt = ""
	}

	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: s.config.Model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   s.config.MaxTokens,
		Temperature: float32(s.config.Temperature),
		TopP:        float32(s.config.TopP),
		Stream:      s.config.Stream,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	return resp.Choices[0].Message.Content, nil
}

func openAIBaseURL(baseURL string) string {
	baseURL = strings.TrimRight(baseURL, "/")
	if strings.HasSuffix(baseURL, "/chat/completions") {
		return strings.TrimSuffix(baseURL, "/chat/completions")
	}
	if strings.HasSuffix(baseURL, "/v1") {
		return baseURL
	}
	return baseURL + "/v1"
}
