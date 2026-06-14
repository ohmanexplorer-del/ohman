package app

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"ohman/config"
	ght "ohman/src/core/github"
)

func publishableGitHubRepos(repos []*ght.RepoData) []*ght.RepoData {
	out := make([]*ght.RepoData, 0, len(repos))
	for _, repo := range repos {
		if repo.Publish || repo.Category == "" {
			out = append(out, repo)
		}
	}
	return out
}

func splitRepoFullName(fullName string) (string, string, bool) {
	parts := strings.SplitN(fullName, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

func (a *App) broadcastGitHubFindings(repos []*ght.RepoData) {
	if a.Telegram == nil || len(repos) == 0 {
		return
	}

	limit := 5
	if len(repos) < limit {
		limit = len(repos)
	}

	lines := []string{
		fmt.Sprintf("Ohman menemukan %d repo menarik:", len(repos)),
	}
	for i := 0; i < limit; i++ {
		repo := repos[i]
		desc := strings.TrimSpace(repo.Description)
		if desc == "" {
			desc = "tanpa deskripsi"
		}
		lines = append(lines, fmt.Sprintf("%d. %s - %s\nhttps://github.com/%s", i+1, repo.FullName, desc, repo.FullName))
	}
	if len(repos) > limit {
		lines = append(lines, fmt.Sprintf("dan %d repo lain sudah ditulis ke README.", len(repos)-limit))
	}
	if err := a.Telegram.Broadcast(context.Background(), strings.Join(lines, "\n\n")); err != nil {
		a.botMu.Lock()
		a.botLastError = err.Error()
		a.botMu.Unlock()
	}
}

func (a *App) notifyActivity(message string) {
	if a == nil || a.Telegram == nil || a.Config == nil || !a.Config.Telegram.ActivityReports || strings.TrimSpace(message) == "" {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.Telegram.Broadcast(ctx, message); err != nil {
			log.Printf("Telegram activity report failed: %v", err)
		}
	}()
}

func (a *App) SetConfigValue(key, value string) error {
	if key = strings.TrimSpace(key); key == "" {
		return fmt.Errorf("config key is required")
	}
	if err := a.DB.Conn().Exec(
		`INSERT INTO configurations (key, value, updated_at)
		 VALUES (?, ?, ?)
		 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at`,
		key,
		value,
		time.Now(),
	).Error; err != nil {
		return err
	}
	cfg, err := config.LoadFromDB(a.DB)
	if err != nil {
		return err
	}
	a.Config = cfg
	a.githubEnabled = cfg.GitHub.Token != ""
	return nil
}

func (a *App) GetConfigValue(key string) (string, bool, error) {
	var row struct {
		Value string
	}
	tx := a.DB.Conn().Table("configurations").Select("value").Where("key = ?", key).Limit(1).Find(&row)
	if tx.Error != nil {
		return "", false, tx.Error
	}
	if tx.RowsAffected == 0 {
		return "", false, nil
	}
	return row.Value, true, nil
}

func (a *App) ListConfigValues(prefix string, limit int) ([]ConfigEntry, error) {
	if limit <= 0 {
		limit = 20
	}
	rows := []ConfigEntry{}
	q := a.DB.Conn().Table("configurations").Select("key, value").Order("key asc").Limit(limit)
	if prefix != "" {
		q = q.Where("key LIKE ?", prefix+"%")
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

type ConfigEntry struct {
	Key   string
	Value string
}
