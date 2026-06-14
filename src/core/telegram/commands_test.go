package telegram

import (
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeController struct {
	status  Status
	started bool
	stopped bool
	err     error
	reports bool
	configs map[string]string
	ran     int
	publish bool
}

func (c *fakeController) StartBot() error {
	c.started = true
	if c.err != nil {
		return c.err
	}
	c.status.Running = true
	now := time.Unix(1, 0)
	c.status.StartedAt = &now
	return nil
}

func (c *fakeController) StopBot() error {
	c.stopped = true
	if c.err != nil {
		return c.err
	}
	c.status.Running = false
	now := time.Unix(2, 0)
	c.status.StoppedAt = &now
	return nil
}

func (c *fakeController) BotStatus() Status {
	return c.status
}

func (c *fakeController) SetActivityReports(enabled bool) error {
	if c.err != nil {
		return c.err
	}
	c.reports = enabled
	return nil
}

func (c *fakeController) RunDiscoveryNow(limit int) error {
	if c.err != nil {
		return c.err
	}
	c.ran = limit
	return nil
}

func (c *fakeController) PublishLibraryNow() error {
	if c.err != nil {
		return c.err
	}
	c.publish = true
	return nil
}

func (c *fakeController) RescoreNow(limit int) error {
	if c.err != nil {
		return c.err
	}
	return nil
}

func (c *fakeController) SetConfigValue(key, value string) error {
	if c.err != nil {
		return c.err
	}
	if c.configs == nil {
		c.configs = map[string]string{}
	}
	c.configs[key] = value
	return nil
}

func (c *fakeController) GetConfigValue(key string) (string, bool, error) {
	if c.err != nil {
		return "", false, c.err
	}
	value, ok := c.configs[key]
	return value, ok, nil
}

func (c *fakeController) ListConfigValues(prefix string, limit int) ([]ConfigEntry, error) {
	if c.err != nil {
		return nil, c.err
	}
	var out []ConfigEntry
	for key, value := range c.configs {
		if prefix == "" || strings.HasPrefix(key, prefix) {
			out = append(out, ConfigEntry{Key: key, Value: value})
		}
	}
	return out, nil
}

func TestNormalizeCommand(t *testing.T) {
	if got := normalizeCommand(" /start@OhmanBot now "); got != "/start" {
		t.Fatalf("expected normalized command, got %q", got)
	}
}

func TestExecuteStartCommand(t *testing.T) {
	controller := &fakeController{}
	service := &Service{controller: controller}

	reply := service.executeCommand("/start")
	if !controller.started {
		t.Fatal("expected start to be called")
	}
	if !strings.Contains(reply, "bot started") || !strings.Contains(reply, "status: running") {
		t.Fatalf("unexpected reply: %s", reply)
	}
}

func TestExecuteStopCommand(t *testing.T) {
	controller := &fakeController{status: Status{Running: true}}
	service := &Service{controller: controller}

	reply := service.executeCommand("/stop")
	if !controller.stopped {
		t.Fatal("expected stop to be called")
	}
	if !strings.Contains(reply, "bot stopped") || !strings.Contains(reply, "status: stopped") {
		t.Fatalf("unexpected reply: %s", reply)
	}
}

func TestExecuteRestartCommand(t *testing.T) {
	controller := &fakeController{status: Status{Running: true}}
	service := &Service{controller: controller}

	reply := service.executeCommand("/restart")
	if !controller.stopped || !controller.started {
		t.Fatal("expected stop and start to be called")
	}
	if !strings.Contains(reply, "bot restarted") || !strings.Contains(reply, "status: running") {
		t.Fatalf("unexpected reply: %s", reply)
	}
}

func TestExecuteUnknownCommand(t *testing.T) {
	service := &Service{controller: &fakeController{}}
	reply := service.executeCommand("/unknown")
	if !strings.Contains(reply, "unknown command") {
		t.Fatalf("unexpected reply: %s", reply)
	}
}

func TestExecuteStartCommandError(t *testing.T) {
	service := &Service{controller: &fakeController{err: errors.New("boom")}}
	reply := service.executeCommand("/start")
	if !strings.Contains(reply, "start failed: boom") {
		t.Fatalf("unexpected reply: %s", reply)
	}
}

func TestExecuteReportsOnCommand(t *testing.T) {
	controller := &fakeController{}
	service := &Service{controller: controller}
	reply := service.executeCommand("/reports_on")
	if !controller.reports || !service.cfg.ActivityReports {
		t.Fatal("expected reports to be enabled")
	}
	if !strings.Contains(reply, "activity reports enabled") {
		t.Fatalf("unexpected reply: %s", reply)
	}
}

func TestExecuteSetLimitCommand(t *testing.T) {
	controller := &fakeController{}
	service := &Service{controller: controller}
	reply := service.executeCommand("/set_limit 7")
	if controller.configs["exploration.max_repos_per_run"] != "7" {
		t.Fatal("expected config to be updated")
	}
	if !strings.Contains(reply, "exploration.max_repos_per_run=7") {
		t.Fatalf("unexpected reply: %s", reply)
	}
}

func TestExecuteRunOnceCommand(t *testing.T) {
	controller := &fakeController{}
	service := &Service{controller: controller}
	reply := service.executeCommand("/run_once 3")
	if controller.ran != 3 {
		t.Fatalf("expected run limit 3, got %d", controller.ran)
	}
	if !strings.Contains(reply, "discovery run requested") {
		t.Fatalf("unexpected reply: %s", reply)
	}
}

func TestExecuteSetIntervalCommand(t *testing.T) {
	controller := &fakeController{}
	service := &Service{controller: controller}
	reply := service.executeCommand("/set_interval 0 */20 * * * *")
	if controller.configs["schedule.exploration_interval"] != "0 */20 * * * *" {
		t.Fatal("expected cron expression to be preserved")
	}
	if !strings.Contains(reply, "schedule.exploration_interval=0 */20 * * * *") {
		t.Fatalf("unexpected reply: %s", reply)
	}
}

func TestExecuteSetQueriesCommand(t *testing.T) {
	controller := &fakeController{}
	service := &Service{controller: controller}
	reply := service.executeCommand(`/set_queries ["stars:10..100 pushed:>2026-01-01 archived:false fork:false"]`)
	if controller.configs["github.search_queries"] == "" {
		t.Fatal("expected search queries to be updated")
	}
	if !strings.Contains(reply, "github.search_queries updated") {
		t.Fatalf("unexpected reply: %s", reply)
	}
}

func TestExecuteSetQueriesCommandRejectsInvalidJSON(t *testing.T) {
	controller := &fakeController{}
	service := &Service{controller: controller}
	reply := service.executeCommand("/set_queries stars:10")
	if controller.configs["github.search_queries"] != "" {
		t.Fatal("expected invalid JSON to be rejected")
	}
	if !strings.Contains(reply, "JSON string array") {
		t.Fatalf("unexpected reply: %s", reply)
	}
}

func TestMaskConfigValue(t *testing.T) {
	if got := maskConfigValue("telegram.token", "abc"); got != "set" {
		t.Fatalf("expected masked token, got %q", got)
	}
}
