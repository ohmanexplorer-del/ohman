package explorer

import (
	"context"
	"fmt"
	"strings"

	"ohman/src/core/prompts"
)

type AINavigator struct {
	aiService AIContentService
	config    AINavigatorConfig
}

type AIContentService interface {
	GenerateContent(ctx context.Context, prompt string) (string, error)
}

type AINavigatorConfig struct {
	MaxTokens     int
	Temperature   float32
	DecisionModel string
}

func NewAINavigator(aiService AIContentService, config AINavigatorConfig) *AINavigator {
	return &AINavigator{
		aiService: aiService,
		config:    config,
	}
}

func (n *AINavigator) ShouldNavigate(ctx context.Context, currentURL, targetURL string) (bool, string, error) {
	prompt, err := prompts.RenderNavigator(prompts.NavigatorPromptData{
		CurrentURL: currentURL,
		TargetURL:  targetURL,
	})
	if err != nil {
		return false, "", fmt.Errorf("failed to render navigator prompt: %w", err)
	}

	response, err := n.aiService.GenerateContent(ctx, prompt)
	if err != nil {
		return false, "", fmt.Errorf("failed to generate AI decision: %w", err)
	}

	response = strings.TrimSpace(response)

	if strings.HasPrefix(strings.ToUpper(response), "YES") {
		reason := strings.TrimPrefix(response, "YES:")
		reason = strings.TrimPrefix(reason, "yes:")
		reason = strings.TrimSpace(reason)
		if reason == "" {
			reason = "AI approved navigation"
		}
		return true, reason, nil
	} else if strings.HasPrefix(strings.ToUpper(response), "NO") {
		reason := strings.TrimPrefix(response, "NO:")
		reason = strings.TrimPrefix(reason, "no:")
		reason = strings.TrimSpace(reason)
		if reason == "" {
			reason = "AI declined navigation"
		}
		return false, reason, nil
	}

	return false, "AI decision unclear", nil
}

func (n *AINavigator) GenerateNextAction(ctx context.Context, currentPage PageData) (string, error) {
	prompt, err := prompts.RenderExplorer(prompts.ExplorerPromptData{
		URL:            currentPage.URL,
		Title:          currentPage.Title,
		ContentPreview: truncateContent(currentPage.Content, 500),
	})
	if err != nil {
		return "CONTINUE", fmt.Errorf("failed to render explorer prompt: %w", err)
	}

	response, err := n.aiService.GenerateContent(ctx, prompt)
	if err != nil {
		return "CONTINUE", fmt.Errorf("failed to generate action: %w", err)
	}

	response = strings.TrimSpace(strings.ToUpper(response))

	validActions := map[string]bool{
		"CONTINUE":  true,
		"BACKTRACK": true,
		"EXPLORE":   true,
		"ANALYZE":   true,
		"STOP":      true,
	}

	if validActions[response] {
		return response, nil
	}

	return "CONTINUE", nil
}

func truncateContent(content string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}
	return content[:maxLen] + "..."
}
