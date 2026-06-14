package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

func NewService(cfg Config, controller BotController) *Service {
	ctx, cancel := context.WithCancel(context.Background())
	allowed := make(map[int64]bool, len(cfg.AllowedChatIDs))
	for _, id := range cfg.AllowedChatIDs {
		allowed[id] = true
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 3 * time.Second
	}
	return &Service{
		cfg:        cfg,
		controller: controller,
		http:       &http.Client{Timeout: 35 * time.Second},
		allowed:    allowed,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (s *Service) Start() {
	if s == nil || !s.cfg.Enabled || s.cfg.Token == "" || s.controller == nil {
		return
	}
	if err := s.skipPendingUpdates(); err != nil {
		log.Printf("Telegram pending update sync failed: %v", err)
	}
	go s.loop()
}

func (s *Service) Stop() {
	if s == nil || s.cancel == nil {
		return
	}
	s.cancel()
}

func (s *Service) Broadcast(ctx context.Context, text string) error {
	if s == nil || !s.cfg.Enabled || s.cfg.Token == "" || strings.TrimSpace(text) == "" {
		return nil
	}
	if s.cfg.ChannelChatID != 0 {
		return s.SendMessage(ctx, s.cfg.ChannelChatID, text)
	}
	for _, chatID := range s.cfg.AllowedChatIDs {
		if err := s.SendMessage(ctx, chatID, text); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) SendMessage(ctx context.Context, chatID int64, text string) error {
	payload, err := json.Marshal(sendMessageRequest{
		ChatID:                chatID,
		Text:                  text,
		DisableWebPagePreview: false,
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.apiURL("sendMessage"), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("telegram sendMessage returned status %d", resp.StatusCode)
	}
	return nil
}

func (s *Service) loop() {
	ticker := time.NewTicker(s.cfg.PollInterval)
	defer ticker.Stop()
	for {
		if err := s.pollOnce(); err != nil {
			log.Printf("Telegram polling error: %v", err)
		}
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *Service) pollOnce() error {
	updates, err := s.getUpdates()
	if err != nil {
		return err
	}
	for _, update := range updates {
		if update.UpdateID >= s.offset {
			s.offset = update.UpdateID + 1
		}
		if update.Message == nil || strings.TrimSpace(update.Message.Text) == "" {
			continue
		}
		if !s.allowedChat(update.Message.Chat.ID) {
			continue
		}
		reply := s.executeCommand(update.Message.Text)
		if reply == "" {
			continue
		}
		if err := s.SendMessage(s.ctx, update.Message.Chat.ID, reply); err != nil {
			log.Printf("Telegram reply failed: %v", err)
		}
	}
	return nil
}

func (s *Service) getUpdates() ([]telegramUpdate, error) {
	return s.fetchUpdates(25)
}

func (s *Service) fetchUpdates(timeout int) ([]telegramUpdate, error) {
	payload, err := json.Marshal(getUpdatesRequest{
		Offset:  s.offset,
		Timeout: timeout,
		AllowedUpdates: []string{
			"message",
		},
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(s.ctx, http.MethodPost, s.apiURL("getUpdates"), bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("telegram getUpdates returned status %d", resp.StatusCode)
	}

	var out getUpdatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if !out.OK {
		return nil, fmt.Errorf("telegram getUpdates returned ok=false")
	}
	return out.Result, nil
}

func (s *Service) skipPendingUpdates() error {
	updates, err := s.fetchUpdates(0)
	if err != nil {
		return err
	}
	for _, update := range updates {
		if update.UpdateID >= s.offset {
			s.offset = update.UpdateID + 1
		}
	}
	return nil
}

func (s *Service) allowedChat(chatID int64) bool {
	return len(s.allowed) > 0 && s.allowed[chatID]
}

func (s *Service) apiURL(method string) string {
	return "https://api.telegram.org/bot" + s.cfg.Token + "/" + method
}
