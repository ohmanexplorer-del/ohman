package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DB struct {
	conn *gorm.DB
}

type repoRow struct {
	ID            uint      `gorm:"primaryKey"`
	GithubID      int64     `gorm:"uniqueIndex"`
	FullName      string    `gorm:"size:512;index"`
	Owner         string    `gorm:"size:256;index"`
	Description   string    `gorm:"type:text"`
	Stars         int       `gorm:"default:0"`
	Forks         int       `gorm:"default:0"`
	Language      string    `gorm:"size:128"`
	Topics        string    `gorm:"type:text"`
	License       string    `gorm:"size:128"`
	DefaultBranch string    `gorm:"size:128"`
	Homepage      string    `gorm:"size:1024"`
	AIScore       float64   `gorm:"default:0"`
	AINotes       string    `gorm:"type:text"`
	ReadmeSummary string    `gorm:"type:text"`
	Category      string    `gorm:"size:128;index"`
	ProjectType   string    `gorm:"size:128"`
	Novelty       float64   `gorm:"default:0"`
	Maturity      float64   `gorm:"default:0"`
	SmallRepoFit  float64   `gorm:"default:0"`
	Strengths     string    `gorm:"type:text"`
	Weaknesses    string    `gorm:"type:text"`
	Publish       bool      `gorm:"default:true;index"`
	VisitedAt     time.Time `gorm:"index"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	PushedAt      time.Time
}

func (repoRow) TableName() string { return "repos" }

type RepoData struct {
	GithubID      int64     `json:"github_id"`
	FullName      string    `json:"full_name"`
	Owner         string    `json:"owner"`
	Description   string    `json:"description"`
	Stars         int       `json:"stars"`
	Forks         int       `json:"forks"`
	Language      string    `json:"language"`
	Topics        []string  `json:"topics"`
	License       string    `json:"license"`
	DefaultBranch string    `json:"default_branch"`
	Homepage      string    `json:"homepage"`
	VisitedAt     time.Time `json:"visited_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	PushedAt      time.Time `json:"pushed_at"`
	AIScore       float64   `json:"ai_score"`
	AINotes       string    `json:"ai_notes"`
	ReadmeSummary string    `json:"readme_summary"`
	Category      string    `json:"category"`
	ProjectType   string    `json:"project_type"`
	Novelty       float64   `json:"novelty"`
	Maturity      float64   `json:"maturity"`
	SmallRepoFit  float64   `json:"small_repo_fit"`
	Strengths     []string  `json:"strengths"`
	Weaknesses    []string  `json:"weaknesses"`
	Publish       bool      `json:"publish"`
}

type UserData struct {
	GithubID    int64     `json:"github_id"`
	Username    string    `json:"username"`
	Name        string    `json:"name"`
	Bio         string    `json:"bio"`
	Followers   int       `json:"followers"`
	Following   int       `json:"following"`
	PublicRepos int       `json:"public_repos"`
	VisitedAt   time.Time `json:"visited_at"`
}

type userRow struct {
	ID          uint      `gorm:"primaryKey"`
	GithubID    int64     `gorm:"uniqueIndex"`
	Username    string    `gorm:"size:256;index"`
	Name        string    `gorm:"size:256"`
	Bio         string    `gorm:"type:text"`
	Followers   int       `gorm:"default:0"`
	Following   int       `gorm:"default:0"`
	PublicRepos int       `gorm:"default:0"`
	VisitedAt   time.Time `gorm:"index"`
}

func (userRow) TableName() string { return "users" }

type sessionRow struct {
	ID            int64      `gorm:"primaryKey"`
	Strategy      string     `gorm:"size:128;not null;default:serendipity"`
	StartedAt     time.Time  `gorm:"default:CURRENT_TIMESTAMP"`
	EndedAt       *time.Time `gorm:"default:null"`
	ReposVisited  int        `gorm:"default:0"`
	DecisionsMade int        `gorm:"default:0"`
	AITokensUsed  int        `gorm:"default:0"`
}

func (sessionRow) TableName() string { return "sessions" }

type edgeRow struct {
	ID           uint      `gorm:"primaryKey"`
	FromUser     string    `gorm:"size:256;not null;uniqueIndex:idx_edge_uniq,priority:1"`
	ToUser       string    `gorm:"size:256;not null;uniqueIndex:idx_edge_uniq,priority:2"`
	Relation     string    `gorm:"size:128;not null;uniqueIndex:idx_edge_uniq,priority:3"`
	DiscoveredAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

func (edgeRow) TableName() string { return "social_graph" }

type crawlQueueRow struct {
	ID           uint       `gorm:"primaryKey"`
	Username     string     `gorm:"size:256;uniqueIndex"`
	Status       string     `gorm:"size:32;index"`
	SessionID    *int64     `gorm:"index;default:null"`
	DiscoveredAt time.Time  `gorm:"autoCreateTime"`
	ProcessedAt  *time.Time `gorm:"default:null"`
}

func (crawlQueueRow) TableName() string { return "crawl_queue" }

type configRow struct {
	ID        uint      `gorm:"primaryKey"`
	Key       string    `gorm:"size:256;uniqueIndex"`
	Value     string    `gorm:"type:text"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (configRow) TableName() string { return "configurations" }

type accountRow struct {
	ID        uint      `gorm:"primaryKey"`
	Provider  string    `gorm:"size:64;index"`
	Name      string    `gorm:"size:256"`
	Token     string    `gorm:"size:1024"`
	Username  string    `gorm:"size:256"`
	Extra     string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (accountRow) TableName() string { return "accounts" }

type githubHTTPCacheRow struct {
	ID        uint      `gorm:"primaryKey"`
	Key       string    `gorm:"size:1024;uniqueIndex"`
	ETag      string    `gorm:"size:512"`
	Body      string    `gorm:"type:text"`
	Status    int       `gorm:"default:200"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (githubHTTPCacheRow) TableName() string { return "github_http_cache" }

func New(dsn string) (*DB, error) {
	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.Migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}
	log.Printf("database connected and migrated")
	return db, nil
}

func (db *DB) Close() error {
	sqlDB, err := db.conn.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	return sqlDB.Close()
}

func (db *DB) Migrate() error {
	return db.conn.AutoMigrate(
		&repoRow{},
		&userRow{},
		&sessionRow{},
		&edgeRow{},
		&crawlQueueRow{},
		&configRow{},
		&accountRow{},
		&githubHTTPCacheRow{},
	)
}

func (db *DB) Conn() *gorm.DB {
	return db.conn
}

func (db *DB) GetHTTPCache(key string) (string, []byte, int, bool, error) {
	var rows []githubHTTPCacheRow
	if err := db.conn.Where("key = ?", key).Limit(1).Find(&rows).Error; err != nil {
		return "", nil, 0, false, err
	}
	if len(rows) == 0 {
		return "", nil, 0, false, nil
	}
	row := rows[0]
	return row.ETag, []byte(row.Body), row.Status, true, nil
}

func (db *DB) SetHTTPCache(key, etag string, body []byte, status int) error {
	if etag == "" || len(body) == 0 {
		return nil
	}
	row := githubHTTPCacheRow{
		Key:    key,
		ETag:   etag,
		Body:   string(body),
		Status: status,
	}
	return db.conn.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		UpdateAll: true,
	}).Create(&row).Error
}

func (db *DB) UpsertRepo(r *RepoData) error {
	topicsJSON, err := json.Marshal(r.Topics)
	if err != nil {
		return fmt.Errorf("failed to marshal topics: %w", err)
	}
	strengthsJSON, err := json.Marshal(r.Strengths)
	if err != nil {
		return fmt.Errorf("failed to marshal strengths: %w", err)
	}
	weaknessesJSON, err := json.Marshal(r.Weaknesses)
	if err != nil {
		return fmt.Errorf("failed to marshal weaknesses: %w", err)
	}
	publish := r.Publish || r.AIScore == 0

	row := repoRow{
		GithubID:      r.GithubID,
		FullName:      r.FullName,
		Owner:         r.Owner,
		Description:   r.Description,
		Stars:         r.Stars,
		Forks:         r.Forks,
		Language:      r.Language,
		Topics:        string(topicsJSON),
		License:       r.License,
		DefaultBranch: r.DefaultBranch,
		Homepage:      r.Homepage,
		AIScore:       r.AIScore,
		AINotes:       r.AINotes,
		ReadmeSummary: r.ReadmeSummary,
		Category:      r.Category,
		ProjectType:   r.ProjectType,
		Novelty:       r.Novelty,
		Maturity:      r.Maturity,
		SmallRepoFit:  r.SmallRepoFit,
		Strengths:     string(strengthsJSON),
		Weaknesses:    string(weaknessesJSON),
		Publish:       publish,
		VisitedAt:     r.VisitedAt,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
		PushedAt:      r.PushedAt,
	}

	result := db.conn.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "github_id"}},
		UpdateAll: true,
	}).Create(&row)
	if result.Error != nil {
		return fmt.Errorf("failed to upsert repo %s: %w", r.FullName, result.Error)
	}
	return nil
}

func (db *DB) UpsertUser(u *UserData) error {
	row := userRow{
		GithubID:    u.GithubID,
		Username:    u.Username,
		Name:        u.Name,
		Bio:         u.Bio,
		Followers:   u.Followers,
		Following:   u.Following,
		PublicRepos: u.PublicRepos,
		VisitedAt:   u.VisitedAt,
	}

	result := db.conn.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "github_id"}},
		UpdateAll: true,
	}).Create(&row)
	if result.Error != nil {
		return fmt.Errorf("failed to upsert user %s: %w", u.Username, result.Error)
	}
	return nil
}

func (db *DB) InsertEdge(from, to, relation string) error {
	row := edgeRow{
		FromUser:     from,
		ToUser:       to,
		Relation:     relation,
		DiscoveredAt: time.Now(),
	}

	result := db.conn.Clauses(clause.OnConflict{DoNothing: true}).Create(&row)
	if result.Error != nil {
		return fmt.Errorf("failed to insert edge %s->%s: %w", from, to, result.Error)
	}
	return nil
}

func (db *DB) EnqueueUser(username string) error {
	row := crawlQueueRow{
		Username:    username,
		Status:      "pending",
		ProcessedAt: nil,
	}

	result := db.conn.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "username"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"status":       "pending",
			"session_id":   nil,
			"processed_at": nil,
		}),
	}).Create(&row)
	if result.Error != nil {
		return fmt.Errorf("failed to enqueue user %s: %w", username, result.Error)
	}
	return nil
}

func (db *DB) EnqueueUserWithSession(username string, sessionID int64) error {
	row := crawlQueueRow{
		Username:    username,
		Status:      "pending",
		SessionID:   &sessionID,
		ProcessedAt: nil,
	}

	result := db.conn.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "username"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"status":       "pending",
			"session_id":   &sessionID,
			"processed_at": nil,
		}),
	}).Create(&row)
	if result.Error != nil {
		return fmt.Errorf("failed to enqueue user %s: %w", username, result.Error)
	}
	return nil
}

func (db *DB) PopNextUser(sessionID *int64) (string, bool, error) {
	tx := db.conn.Begin()
	if tx.Error != nil {
		return "", false, tx.Error
	}

	var row crawlQueueRow
	q := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).Where("status = ?", "pending")
	if sessionID != nil {
		q = q.Where("session_id = ?", *sessionID)
	}
	err := q.Order("id asc").First(&row).Error
	if err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return "", false, nil
		}
		return "", false, err
	}

	if err := tx.Model(&crawlQueueRow{}).Where("id = ? AND status = ?", row.ID, "pending").Updates(map[string]interface{}{"status": "processing"}).Error; err != nil {
		tx.Rollback()
		return "", false, err
	}
	if err := tx.Commit().Error; err != nil {
		return "", false, err
	}

	return row.Username, true, nil
}

func (db *DB) IsUserQueued(username string) bool {
	var count int64
	db.conn.Model(&crawlQueueRow{}).Where("username = ? AND status IN ?", username, []string{"pending", "processing"}).Count(&count)
	return count > 0
}

func (db *DB) GetQueuedUsers() ([]string, error) {
	var rows []crawlQueueRow
	if err := db.conn.Where("status IN ?", []string{"pending", "processing"}).Order("id asc").Find(&rows).Error; err != nil {
		return nil, err
	}

	usernames := make([]string, 0, len(rows))
	for _, row := range rows {
		usernames = append(usernames, row.Username)
	}
	return usernames, nil
}

func (db *DB) MarkUserProcessing(username string) error {
	return db.conn.Model(&crawlQueueRow{}).
		Where("username = ? AND status = ?", username, "pending").
		Updates(map[string]interface{}{"status": "processing"}).
		Error
}

func (db *DB) MarkUserProcessed(username string) error {
	now := time.Now()
	return db.conn.Model(&crawlQueueRow{}).
		Where("username = ?", username).
		Updates(map[string]interface{}{"status": "done", "processed_at": &now}).
		Error
}

func (db *DB) CreateSession(strategy string) (int64, error) {
	row := sessionRow{
		Strategy:  strategy,
		StartedAt: time.Now(),
	}
	if err := db.conn.Create(&row).Error; err != nil {
		return 0, fmt.Errorf("failed to create session: %w", err)
	}
	return row.ID, nil
}

func (db *DB) UpdateSession(id int64, reposVisited, decisionsMade, aiTokensUsed int) error {
	now := time.Now()
	result := db.conn.Model(&sessionRow{}).Where("id = ?", id).Updates(map[string]interface{}{
		"ended_at":       &now,
		"repos_visited":  reposVisited,
		"decisions_made": decisionsMade,
		"ai_tokens_used": aiTokensUsed,
	})
	if result.Error != nil {
		return fmt.Errorf("failed to update session %d: %w", id, result.Error)
	}
	return nil
}

func (db *DB) HasRepo(fullName string) bool {
	var count int64
	db.conn.Model(&repoRow{}).Where("full_name = ?", fullName).Count(&count)
	return count > 0
}

func (db *DB) HasUser(username string) bool {
	var count int64
	db.conn.Model(&userRow{}).Where("username = ?", username).Count(&count)
	return count > 0
}

func (db *DB) GetVisitedRepos(limit int) ([]*RepoData, error) {
	var rows []repoRow
	if err := db.conn.Order("stars DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("failed to query repos: %w", err)
	}

	repos := make([]*RepoData, 0, len(rows))
	for _, r := range rows {
		var topics []string
		var strengths []string
		var weaknesses []string
		if r.Topics != "" {
			if err := json.Unmarshal([]byte(r.Topics), &topics); err != nil {
				topics = nil
			}
		}
		if r.Strengths != "" {
			if err := json.Unmarshal([]byte(r.Strengths), &strengths); err != nil {
				strengths = nil
			}
		}
		if r.Weaknesses != "" {
			if err := json.Unmarshal([]byte(r.Weaknesses), &weaknesses); err != nil {
				weaknesses = nil
			}
		}
		repos = append(repos, &RepoData{
			GithubID:      r.GithubID,
			FullName:      r.FullName,
			Owner:         r.Owner,
			Description:   r.Description,
			Stars:         r.Stars,
			Forks:         r.Forks,
			Language:      r.Language,
			Topics:        topics,
			License:       r.License,
			DefaultBranch: r.DefaultBranch,
			Homepage:      r.Homepage,
			AIScore:       r.AIScore,
			AINotes:       r.AINotes,
			ReadmeSummary: r.ReadmeSummary,
			Category:      r.Category,
			ProjectType:   r.ProjectType,
			Novelty:       r.Novelty,
			Maturity:      r.Maturity,
			SmallRepoFit:  r.SmallRepoFit,
			Strengths:     strengths,
			Weaknesses:    weaknesses,
			Publish:       r.Publish,
			VisitedAt:     r.VisitedAt,
			CreatedAt:     r.CreatedAt,
			UpdatedAt:     r.UpdatedAt,
			PushedAt:      r.PushedAt,
		})
	}
	return repos, nil
}

func (db *DB) Stats() (map[string]int, error) {
	stats := make(map[string]int)
	tables := []struct {
		model interface{}
		name  string
	}{
		{&repoRow{}, "repos"},
		{&userRow{}, "users"},
		{&sessionRow{}, "sessions"},
		{&edgeRow{}, "social_graph"},
		{&crawlQueueRow{}, "crawl_queue"},
		{&configRow{}, "configurations"},
		{&accountRow{}, "accounts"},
		{&githubHTTPCacheRow{}, "github_http_cache"},
	}

	for _, t := range tables {
		var count int64
		if err := db.conn.Model(t.model).Count(&count).Error; err != nil {
			return nil, fmt.Errorf("failed to count %s: %w", t.name, err)
		}
		stats[t.name] = int(count)
	}
	return stats, nil
}
