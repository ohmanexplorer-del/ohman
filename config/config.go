package config

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	storage "ohman/src/core/db"

	"github.com/spf13/viper"
)

type Config struct {
	Crawler      CrawlerConfig      `mapstructure:"crawler"`
	AI           AIConfig           `mapstructure:"ai"`
	API          APIConfig          `mapstructure:"api"`
	Telegram     TelegramConfig     `mapstructure:"telegram"`
	Schedule     ScheduleConfig     `mapstructure:"schedule"`
	Log          LogConfig          `mapstructure:"log"`
	GitHub       GitHubConfig       `mapstructure:"github"`
	Exploration  ExplorationConfig  `mapstructure:"exploration"`
	Contribution ContributionConfig `mapstructure:"contribution"`
	Storage      StorageConfig      `mapstructure:"storage"`
}

type CrawlerConfig struct {
	UserAgent         string        `mapstructure:"user_agent"`
	MaxDepth          int           `mapstructure:"max_depth"`
	MaxPages          int           `mapstructure:"max_pages"`
	Delay             time.Duration `mapstructure:"delay"`
	AllowedDomains    []string      `mapstructure:"allowed_domains"`
	DisallowedDomains []string      `mapstructure:"disallowed_domains"`
	FollowLinks       bool          `mapstructure:"follow_links"`
	RespectRobotsTxt  bool          `mapstructure:"respect_robots_txt"`
	Parallelism       int           `mapstructure:"parallelism"`
	InitialURLs       []string      `mapstructure:"initial_urls"`
}

type AIConfig struct {
	Provider    string  `mapstructure:"provider"`
	BaseURL     string  `mapstructure:"base_url"`
	APIKey      string  `mapstructure:"api_key"`
	Model       string  `mapstructure:"model"`
	MaxTokens   int     `mapstructure:"max_tokens"`
	Temperature float64 `mapstructure:"temperature"`
	TopP        float64 `mapstructure:"top_p"`
	Stream      bool    `mapstructure:"stream"`
}

type APIConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Port    int  `mapstructure:"port"`
}

type TelegramConfig struct {
	Enabled         bool    `mapstructure:"enabled"`
	Token           string  `mapstructure:"token"`
	AllowedChatIDs  []int64 `mapstructure:"allowed_chat_ids"`
	ChannelChatID   int64   `mapstructure:"channel_chat_id"`
	PollInterval    string  `mapstructure:"poll_interval"`
	ActivityReports bool    `mapstructure:"activity_reports"`
}

type ScheduleConfig struct {
	ExplorationInterval string `mapstructure:"exploration_interval"`
	ActiveHours         string `mapstructure:"active_hours"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type GitHubConfig struct {
	Token           string   `mapstructure:"token"`
	BotUsername     string   `mapstructure:"bot_username"`
	BotRepo         string   `mapstructure:"bot_repo"`
	RateLimitBuffer int      `mapstructure:"rate_limit_buffer"`
	SeedUsers       []string `mapstructure:"seed_users"`
	GraphQL         bool     `mapstructure:"graphql"`
	SearchQueries   []string `mapstructure:"search_queries"`
}

type ExplorationConfig struct {
	Mode               string   `mapstructure:"mode"`
	MinStars           int      `mapstructure:"min_stars"`
	Languages          []string `mapstructure:"languages"`
	MaxUsersPerSession int      `mapstructure:"max_users_per_session"`
	ReposPerUser       int      `mapstructure:"repos_per_user"`
	MaxReposPerRun     int      `mapstructure:"max_repos_per_run"`
}

type ContributionConfig struct {
	Enabled         bool     `mapstructure:"enabled"`
	AutoPR          bool     `mapstructure:"auto_pr"`
	Labels          []string `mapstructure:"labels"`
	MaxIssuesPerRun int      `mapstructure:"max_issues_per_run"`
	MaxRepoSizeKB   int      `mapstructure:"max_repo_size_kb"`
	MaxRepoStars    int      `mapstructure:"max_repo_stars"`
	MaxRepoForks    int      `mapstructure:"max_repo_forks"`
	WorkDir         string   `mapstructure:"work_dir"`
}

type StorageConfig struct {
	DatabaseURL string `mapstructure:"database_url"`
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.SetDefault("crawler.user_agent", "ohman-explorer/1.0")
	viper.SetDefault("crawler.max_depth", 3)
	viper.SetDefault("crawler.max_pages", 100)
	viper.SetDefault("crawler.delay", "2s")
	viper.SetDefault("crawler.follow_links", true)
	viper.SetDefault("crawler.respect_robots_txt", true)
	viper.SetDefault("crawler.parallelism", 2)
	viper.SetDefault("crawler.initial_urls", []string{"https://example.com"})
	viper.SetDefault("ai.provider", "genai-gtw")
	viper.SetDefault("ai.base_url", "http://localhost:8080/v1")
	viper.SetDefault("ai.api_key", "")
	viper.SetDefault("ai.model", "combo")
	viper.SetDefault("ai.max_tokens", 300)
	viper.SetDefault("ai.temperature", 0.7)
	viper.SetDefault("ai.top_p", 0.9)
	viper.SetDefault("ai.stream", false)
	viper.SetDefault("api.enabled", false)
	viper.SetDefault("api.port", 8081)
	viper.SetDefault("telegram.enabled", false)
	viper.SetDefault("telegram.poll_interval", "3s")
	viper.SetDefault("telegram.activity_reports", true)
	viper.SetDefault("schedule.exploration_interval", "0 */20 * * * *")
	viper.SetDefault("schedule.active_hours", "09:00-21:00")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("github.rate_limit_buffer", 100)
	viper.SetDefault("github.seed_users", []string{"torvalds", "rakyll", "kelseyhightower"})
	viper.SetDefault("github.graphql", true)
	viper.SetDefault("github.search_queries", defaultGitHubSearchQueries())
	viper.SetDefault("exploration.mode", "serendipity")
	viper.SetDefault("exploration.min_stars", 10)
	viper.SetDefault("exploration.max_users_per_session", 5)
	viper.SetDefault("exploration.repos_per_user", 5)
	viper.SetDefault("exploration.max_repos_per_run", 10)
	viper.SetDefault("contribution.enabled", false)
	viper.SetDefault("contribution.auto_pr", false)
	viper.SetDefault("contribution.labels", []string{"good first issue", "bug"})
	viper.SetDefault("contribution.max_issues_per_run", 2)
	viper.SetDefault("contribution.max_repo_size_kb", 50000)
	viper.SetDefault("contribution.max_repo_stars", 5000)
	viper.SetDefault("contribution.max_repo_forks", 1000)
	viper.SetDefault("contribution.work_dir", "work/contributions")
	viper.SetDefault("storage.database_url", "postgres://ohman:ohman_secret@localhost:5432/ohman?sslmode=disable")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if len(config.Crawler.InitialURLs) == 0 {
		return nil, fmt.Errorf("crawler.initial_urls is required")
	}
	log.Printf("Configuration loaded successfully")
	return &config, nil
}

func LoadFromDB(db *storage.DB) (*Config, error) {
	viper.SetDefault("crawler.user_agent", "ohman-explorer/1.0")
	viper.SetDefault("crawler.max_depth", 3)
	viper.SetDefault("crawler.max_pages", 100)
	viper.SetDefault("crawler.delay", "2s")
	viper.SetDefault("crawler.follow_links", true)
	viper.SetDefault("crawler.respect_robots_txt", true)
	viper.SetDefault("crawler.parallelism", 2)
	viper.SetDefault("crawler.initial_urls", []string{"https://example.com"})
	viper.SetDefault("ai.provider", "genai-gtw")
	viper.SetDefault("ai.base_url", "http://localhost:8080/v1")
	viper.SetDefault("ai.api_key", "")
	viper.SetDefault("ai.model", "combo")
	viper.SetDefault("ai.max_tokens", 300)
	viper.SetDefault("ai.temperature", 0.7)
	viper.SetDefault("ai.top_p", 0.9)
	viper.SetDefault("ai.stream", false)
	viper.SetDefault("api.enabled", false)
	viper.SetDefault("api.port", 8081)
	viper.SetDefault("telegram.enabled", false)
	viper.SetDefault("telegram.poll_interval", "3s")
	viper.SetDefault("telegram.activity_reports", true)
	viper.SetDefault("schedule.exploration_interval", "0 */20 * * * *")
	viper.SetDefault("schedule.active_hours", "09:00-21:00")
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("github.rate_limit_buffer", 100)
	viper.SetDefault("github.seed_users", []string{"torvalds", "rakyll", "kelseyhightower"})
	viper.SetDefault("github.graphql", true)
	viper.SetDefault("github.search_queries", defaultGitHubSearchQueries())
	viper.SetDefault("exploration.mode", "serendipity")
	viper.SetDefault("exploration.min_stars", 10)
	viper.SetDefault("exploration.max_users_per_session", 5)
	viper.SetDefault("exploration.repos_per_user", 5)
	viper.SetDefault("exploration.max_repos_per_run", 10)
	viper.SetDefault("contribution.enabled", false)
	viper.SetDefault("contribution.auto_pr", false)
	viper.SetDefault("contribution.labels", []string{"good first issue", "bug"})
	viper.SetDefault("contribution.max_issues_per_run", 2)
	viper.SetDefault("contribution.max_repo_size_kb", 50000)
	viper.SetDefault("contribution.max_repo_stars", 5000)
	viper.SetDefault("contribution.max_repo_forks", 1000)
	viper.SetDefault("contribution.work_dir", "work/contributions")
	viper.SetDefault("storage.database_url", "postgres://ohman:ohman_secret@localhost:5432/ohman?sslmode=disable")

	rows := []struct {
		Key   string
		Value string
	}{}
	if err := db.Conn().Table("configurations").Select("key, value").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to read configurations from db: %w", err)
	}
	for _, r := range rows {
		viper.Set(r.Key, parseConfigValue(r.Value))
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config from db values: %w", err)
	}

	if len(config.Crawler.InitialURLs) == 0 {
		log.Printf("warning: crawler.initial_urls empty in DB; using defaults")
	}

	return &config, nil
}

func defaultGitHubSearchQueries() []string {
	return []string{
		"stars:10..500 pushed:>2026-01-01 archived:false fork:false",
		"stars:10..200 created:>2026-01-01 pushed:>2026-01-01 archived:false fork:false",
		"topic:ai-agent stars:10..1000 pushed:>2026-01-01 archived:false fork:false",
		"topic:developer-tools stars:10..1000 pushed:>2026-01-01 archived:false fork:false",
		"topic:database stars:10..1000 pushed:>2026-01-01 archived:false fork:false",
		"topic:security stars:10..1000 pushed:>2026-01-01 archived:false fork:false",
		"topic:cli stars:10..1000 pushed:>2026-01-01 archived:false fork:false",
		"topic:web-framework stars:10..1000 pushed:>2026-01-01 archived:false fork:false",
	}
}

func parseConfigValue(value string) any {
	trimmed := strings.TrimSpace(value)
	if strings.HasPrefix(trimmed, "[") {
		var stringsValue []string
		if err := json.Unmarshal([]byte(trimmed), &stringsValue); err == nil {
			return stringsValue
		}
		var int64Value []int64
		if err := json.Unmarshal([]byte(trimmed), &int64Value); err == nil {
			return int64Value
		}
	}
	return value
}
