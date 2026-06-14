package app

import (
	"time"

	"ohman/src/core/telegram"
)

type telegramController struct {
	app *App
}

func (c telegramController) StartBot() error {
	return c.app.StartBot()
}

func (c telegramController) StopBot() error {
	return c.app.StopBot()
}

func (c telegramController) BotStatus() telegram.Status {
	status := c.app.BotStatus()
	return telegram.Status{
		Running:       status.Running,
		GitHubEnabled: status.GitHubEnabled,
		StartedAt:     status.StartedAt,
		StoppedAt:     status.StoppedAt,
		LastError:     status.LastError,
	}
}

func (c telegramController) SetActivityReports(enabled bool) error {
	value := "false"
	if enabled {
		value = "true"
	}
	if err := c.app.DB.Conn().Exec(
		`INSERT INTO configurations (key, value, updated_at)
		 VALUES (?, ?, ?)
		 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = EXCLUDED.updated_at`,
		"telegram.activity_reports",
		value,
		time.Now(),
	).Error; err != nil {
		return err
	}
	if c.app.Config != nil {
		c.app.Config.Telegram.ActivityReports = enabled
	}
	return nil
}

func (c telegramController) RunDiscoveryNow(limit int) error {
	return c.app.RunGitHubDiscoveryNow(limit)
}

func (c telegramController) PublishLibraryNow() error {
	return c.app.PublishLibraryNow()
}

func (c telegramController) RescoreNow(limit int) error {
	return c.app.RescoreNow(limit)
}

func (c telegramController) SetConfigValue(key, value string) error {
	return c.app.SetConfigValue(key, value)
}

func (c telegramController) GetConfigValue(key string) (string, bool, error) {
	return c.app.GetConfigValue(key)
}

func (c telegramController) ListConfigValues(prefix string, limit int) ([]telegram.ConfigEntry, error) {
	rows, err := c.app.ListConfigValues(prefix, limit)
	if err != nil {
		return nil, err
	}
	out := make([]telegram.ConfigEntry, 0, len(rows))
	for _, row := range rows {
		out = append(out, telegram.ConfigEntry{Key: row.Key, Value: row.Value})
	}
	return out, nil
}
