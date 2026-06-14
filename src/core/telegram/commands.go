package telegram

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (s *Service) executeCommand(text string) string {
	command, args := parseCommand(text)
	switch command {
	case "/start":
		if err := s.controller.StartBot(); err != nil {
			return "start failed: " + err.Error()
		}
		return "bot started\n\n" + formatStatus(s.controller.BotStatus())
	case "/stop":
		if err := s.controller.StopBot(); err != nil {
			return "stop failed: " + err.Error()
		}
		return "bot stopped\n\n" + formatStatus(s.controller.BotStatus())
	case "/restart":
		if err := s.controller.StopBot(); err != nil {
			return "restart failed: " + err.Error()
		}
		if err := s.controller.StartBot(); err != nil {
			return "restart failed: " + err.Error()
		}
		return "bot restarted\n\n" + formatStatus(s.controller.BotStatus())
	case "/status":
		return formatStatus(s.controller.BotStatus())
	case "/run_once":
		limit := 0
		if len(args) > 0 {
			value, err := strconv.Atoi(args[0])
			if err != nil || value < 0 {
				return "usage: /run_once [limit]"
			}
			limit = value
		}
		if err := s.controller.RunDiscoveryNow(limit); err != nil {
			return "run_once failed: " + err.Error()
		}
		return "discovery run requested"
	case "/publish":
		if err := s.controller.PublishLibraryNow(); err != nil {
			return "publish failed: " + err.Error()
		}
		return "library publish requested"
	case "/set_limit":
		if len(args) != 1 {
			return "usage: /set_limit <number>"
		}
		value, err := strconv.Atoi(args[0])
		if err != nil || value <= 0 {
			return "limit must be a positive number"
		}
		if err := s.controller.SetConfigValue("exploration.max_repos_per_run", strconv.Itoa(value)); err != nil {
			return "set_limit failed: " + err.Error()
		}
		return fmt.Sprintf("exploration.max_repos_per_run=%d", value)
	case "/set_interval":
		if len(args) == 0 {
			return "usage: /set_interval <cron expression>"
		}
		value := strings.Join(args, " ")
		if err := s.controller.SetConfigValue("schedule.exploration_interval", value); err != nil {
			return "set_interval failed: " + err.Error()
		}
		return "schedule.exploration_interval=" + value
	case "/queries":
		value, found, err := s.controller.GetConfigValue("github.search_queries")
		if err != nil {
			return "queries failed: " + err.Error()
		}
		if !found {
			return "github.search_queries is not set"
		}
		return "github.search_queries:\n" + value
	case "/set_queries":
		if len(args) == 0 {
			return "usage: /set_queries <json array>"
		}
		value := strings.Join(args, " ")
		var queries []string
		if err := json.Unmarshal([]byte(value), &queries); err != nil {
			return "set_queries expects a JSON string array"
		}
		if len(queries) == 0 {
			return "set_queries needs at least one query"
		}
		if err := s.controller.SetConfigValue("github.search_queries", value); err != nil {
			return "set_queries failed: " + err.Error()
		}
		return "github.search_queries updated"
	case "/config_get":
		if len(args) != 1 {
			return "usage: /config_get <key>"
		}
		value, found, err := s.controller.GetConfigValue(args[0])
		if err != nil {
			return "config_get failed: " + err.Error()
		}
		if !found {
			return "config not found"
		}
		return args[0] + "=" + maskConfigValue(args[0], value)
	case "/config_set":
		if len(args) < 2 {
			return "usage: /config_set <key> <value>"
		}
		key := args[0]
		value := strings.Join(args[1:], " ")
		if err := s.controller.SetConfigValue(key, value); err != nil {
			return "config_set failed: " + err.Error()
		}
		return key + " updated"
	case "/config_list":
		prefix := ""
		if len(args) > 0 {
			prefix = args[0]
		}
		rows, err := s.controller.ListConfigValues(prefix, 40)
		if err != nil {
			return "config_list failed: " + err.Error()
		}
		if len(rows) == 0 {
			return "no config found"
		}
		lines := []string{"configs:"}
		for _, row := range rows {
			lines = append(lines, row.Key+"="+maskConfigValue(row.Key, row.Value))
		}
		return strings.Join(lines, "\n")
	case "/reports_on":
		if err := s.controller.SetActivityReports(true); err != nil {
			return "failed to enable reports: " + err.Error()
		}
		s.cfg.ActivityReports = true
		return "activity reports enabled"
	case "/reports_off":
		if err := s.controller.SetActivityReports(false); err != nil {
			return "failed to disable reports: " + err.Error()
		}
		s.cfg.ActivityReports = false
		return "activity reports disabled"
	case "/help":
		return commandHelp()
	default:
		if strings.HasPrefix(command, "/") {
			return "unknown command\n\n" + commandHelp()
		}
		return ""
	}
}

func normalizeCommand(text string) string {
	command, _ := parseCommand(text)
	return command
}

func parseCommand(text string) (string, []string) {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) == 0 {
		return "", nil
	}
	command := fields[0]
	if at := strings.Index(command, "@"); at >= 0 {
		command = command[:at]
	}
	return strings.ToLower(command), fields[1:]
}

func formatStatus(status Status) string {
	state := "stopped"
	if status.Running {
		state = "running"
	}

	lines := []string{
		"status: " + state,
		fmt.Sprintf("github: %t", status.GitHubEnabled),
	}
	if status.StartedAt != nil {
		lines = append(lines, "started_at: "+status.StartedAt.Format(time.RFC3339))
	}
	if status.StoppedAt != nil {
		lines = append(lines, "stopped_at: "+status.StoppedAt.Format(time.RFC3339))
	}
	if status.LastError != "" {
		lines = append(lines, "last_error: "+status.LastError)
	}
	return strings.Join(lines, "\n")
}

func commandHelp() string {
	return strings.Join([]string{
		"available commands:",
		"/start - start bot",
		"/stop - stop bot",
		"/restart - restart bot",
		"/status - show bot status",
		"/run_once [limit] - run discovery now",
		"/publish - republish library",
		"/set_limit <number> - set repos per run",
		"/set_interval <cron> - set schedule",
		"/queries - show GitHub search queries",
		"/set_queries <json array> - replace GitHub search queries",
		"/config_get <key> - show config",
		"/config_set <key> <value> - update config",
		"/config_list [prefix] - list configs",
		"/reports_on - enable activity reports",
		"/reports_off - disable activity reports",
		"/help - show commands",
	}, "\n")
}

func maskConfigValue(key, value string) string {
	key = strings.ToLower(key)
	if strings.Contains(key, "token") || strings.Contains(key, "api_key") || strings.Contains(key, "password") || strings.Contains(key, "secret") {
		if value == "" {
			return ""
		}
		return "set"
	}
	return value
}
