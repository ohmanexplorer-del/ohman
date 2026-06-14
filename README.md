# Ohman Explorer

Autonomous GitHub explorer bot for discovering useful public repositories, evaluating them with AI, and publishing a cumulative project library to a bot-owned GitHub repository.

Ohman uses PostgreSQL, GORM auto migration, GitHub API access, an OpenAI-compatible local AI gateway, and Telegram as the main control plane.

## Quick Start

```bash
docker compose up -d postgres ohman
```

## Control the bot from Telegram:

```text
/start
/stop
/restart
/status
/run_once [limit]
/publish
/set_limit <number>
/set_interval <cron>
/queries
/set_queries <json array>
/config_get <key>
/config_set <key> <value>
/config_list [prefix]
/reports_on
/reports_off
/help
```

## Important config keys:

- `ai.base_url`
- `ai.model`
- `github.token`
- `github.bot_username`
- `github.bot_repo`
- `github.search_queries`
- `schedule.exploration_interval`
- `exploration.max_repos_per_run`
- `telegram.enabled`
- `telegram.allowed_chat_ids`

## Local Development

```bash
go test ./...
go run .
go run ./cmd/publish-library
```
