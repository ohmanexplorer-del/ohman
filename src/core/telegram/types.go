package telegram

import (
	"context"
	"net/http"
	"time"
)

type Config struct {
	Enabled         bool
	Token           string
	AllowedChatIDs  []int64
	ChannelChatID   int64
	PollInterval    time.Duration
	ActivityReports bool
}

type BotController interface {
	StartBot() error
	StopBot() error
	BotStatus() Status
	SetActivityReports(enabled bool) error
	RunDiscoveryNow(limit int) error
	PublishLibraryNow() error
	RescoreNow(limit int) error
	SetConfigValue(key, value string) error
	GetConfigValue(key string) (string, bool, error)
	ListConfigValues(prefix string, limit int) ([]ConfigEntry, error)
}

type Status struct {
	Running       bool
	GitHubEnabled bool
	StartedAt     *time.Time
	StoppedAt     *time.Time
	LastError     string
}

type ConfigEntry struct {
	Key   string
	Value string
}

type Service struct {
	cfg        Config
	controller BotController
	http       httpClient
	offset     int64
	allowed    map[int64]bool
	ctx        context.Context
	cancel     context.CancelFunc
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}
