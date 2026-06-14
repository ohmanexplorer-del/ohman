package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type BotStatus struct {
	Running       bool       `json:"running"`
	GitHubEnabled bool       `json:"github_enabled"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	StoppedAt     *time.Time `json:"stopped_at,omitempty"`
	LastError     string     `json:"last_error,omitempty"`
}

type BotController interface {
	StartBot() error
	StopBot() error
	BotStatus() BotStatus
}

type BotHandler struct {
	bot BotController
}

func NewBotHandler(bot BotController) *BotHandler {
	return &BotHandler{bot: bot}
}

func (h *BotHandler) Status(c *gin.Context) {
	c.JSON(http.StatusOK, h.bot.BotStatus())
}

func (h *BotHandler) Start(c *gin.Context) {
	if err := h.bot.StartBot(); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error(), "status": h.bot.BotStatus()})
		return
	}
	c.JSON(http.StatusOK, h.bot.BotStatus())
}

func (h *BotHandler) Stop(c *gin.Context) {
	if err := h.bot.StopBot(); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error(), "status": h.bot.BotStatus()})
		return
	}
	c.JSON(http.StatusOK, h.bot.BotStatus())
}

func (h *BotHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/bot", h.Status)
	rg.POST("/start", h.Start)
	rg.POST("/stop", h.Stop)
}
