package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AccountHandler struct {
	db *gorm.DB
}

func NewAccountHandler(db *gorm.DB) *AccountHandler {
	return &AccountHandler{db: db}
}

type AccountEntry struct {
	ID        uint      `json:"id"`
	Provider  string    `json:"provider"`
	Name      string    `json:"name"`
	Token     string    `json:"token,omitempty"`
	Username  string    `json:"username"`
	Extra     string    `json:"extra,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateAccountRequest struct {
	Provider string `json:"provider" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Token    string `json:"token" binding:"required"`
	Username string `json:"username"`
	Extra    string `json:"extra"`
}

type UpdateAccountRequest struct {
	Name     string `json:"name"`
	Token    string `json:"token"`
	Username string `json:"username"`
	Extra    string `json:"extra"`
}

func (h *AccountHandler) List(c *gin.Context) {
	query := h.db.Table("accounts").Select("id, provider, name, username, created_at, updated_at")
	if p := c.Query("provider"); p != "" {
		query = query.Where("provider = ?", p)
	}
	var rows []AccountEntry
	if err := query.Order("provider ASC, name ASC").Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rows)
}

func (h *AccountHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var row AccountEntry
	if err := h.db.Table("accounts").Select("id, provider, name, username, extra, created_at, updated_at").Where("id = ?", id).Scan(&row).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "account not found"})
		return
	}
	c.JSON(http.StatusOK, row)
}

func (h *AccountHandler) Create(c *gin.Context) {
	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	row := AccountEntry{
		Provider: req.Provider,
		Name:     req.Name,
		Token:    req.Token,
		Username: req.Username,
		Extra:    req.Extra,
	}

	if err := h.db.Table("accounts").Create(&row).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, row)
}

func (h *AccountHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Token != "" {
		updates["token"] = req.Token
	}
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Extra != "" {
		updates["extra"] = req.Extra
	}
	updates["updated_at"] = time.Now()

	if err := h.db.Table("accounts").Where("id = ?", id).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var row AccountEntry
	h.db.Table("accounts").Select("id, provider, name, username, extra, created_at, updated_at").Where("id = ?", id).Scan(&row)
	c.JSON(http.StatusOK, row)
}

func (h *AccountHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.db.Table("accounts").Where("id = ?", id).Delete(nil).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *AccountHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/accounts", h.List)
	rg.GET("/accounts/:id", h.Get)
	rg.POST("/accounts", h.Create)
	rg.PUT("/accounts/:id", h.Update)
	rg.DELETE("/accounts/:id", h.Delete)
}
