package app

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ohman/config"
	api "ohman/src/controllers"
)

func (a *App) Run() {
	if a.API != nil {
		a.API.Start()
	}
	if a.Telegram != nil {
		a.Telegram.Start()
		log.Println("Telegram control polling started")
		a.notifyActivity("Ohman Telegram control aktif. Command tersedia: /start /stop /status /help")
	}
	if a.API != nil {
		log.Println("Ohman Explorer REST API is running. Use POST /api/v1/start to start the bot.")
	} else {
		log.Println("Ohman Explorer REST API is disabled. Telegram control is the primary control plane.")
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	a.Shutdown()
}

func (a *App) Shutdown() {
	log.Println("Shutting down Ohman Explorer...")
	if err := a.StopBot(); err != nil {
		log.Printf("Bot stop skipped: %v", err)
	}
	if a.API != nil {
		a.API.Shutdown(5 * time.Second)
	}
	if a.Telegram != nil {
		a.Telegram.Stop()
	}
	if a.DB != nil {
		a.DB.Close()
	}
	log.Println("Ohman Explorer stopped.")
}

func (a *App) StartBot() error {
	a.botMu.Lock()
	defer a.botMu.Unlock()

	if a.botRunning {
		return nil
	}

	cfg, err := config.LoadFromDB(a.DB)
	if err != nil {
		a.botLastError = err.Error()
		return err
	}
	a.Config = cfg
	a.githubEnabled = cfg.GitHub.Token != ""

	if err := a.initAI(); err != nil {
		a.botLastError = err.Error()
		return err
	}
	a.initGitHub()
	a.initCollector()
	a.initContributor()
	a.initPublisher()
	if err := a.initWebExplorer(); err != nil {
		a.botLastError = err.Error()
		return err
	}
	a.initScheduler()
	a.Scheduler.Start()

	now := time.Now()
	a.botRunning = true
	a.botStartedAt = &now
	a.botStoppedAt = nil
	a.botLastError = ""
	log.Println("Ohman bot started")
	a.notifyActivity("Ohman bot started. Discovery scheduler aktif.")
	if a.githubEnabled {
		go func() {
			if err := a.runGitHubDiscoveryOnce(a.githubReposPerRunLimit()); err != nil {
				log.Printf("GitHub MVP discovery failed: %v", err)
				a.botMu.Lock()
				a.botLastError = err.Error()
				a.botMu.Unlock()
				a.notifyActivity("GitHub discovery gagal: " + err.Error())
			}
		}()
	}
	return nil
}

func (a *App) StopBot() error {
	a.botMu.Lock()
	defer a.botMu.Unlock()

	if !a.botRunning {
		return nil
	}

	if a.Navigator != nil {
		a.Navigator.Stop()
		a.Navigator = nil
	}
	if a.GitHubCrawl != nil {
		a.GitHubCrawl.Stop()
		a.GitHubCrawl = nil
	}
	if a.Scheduler != nil {
		a.Scheduler.Stop()
		a.Scheduler = nil
	}
	a.Collector = nil
	a.Contributor = nil
	a.Publisher = nil
	a.GitHub = nil
	a.AI = nil

	now := time.Now()
	a.botRunning = false
	a.botStoppedAt = &now
	log.Println("Ohman bot stopped")
	a.notifyActivity("Ohman bot stopped.")
	return nil
}

func (a *App) BotStatus() api.BotStatus {
	a.botMu.Lock()
	defer a.botMu.Unlock()

	return api.BotStatus{
		Running:       a.botRunning,
		GitHubEnabled: a.githubEnabled,
		StartedAt:     a.botStartedAt,
		StoppedAt:     a.botStoppedAt,
		LastError:     a.botLastError,
	}
}
