package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ConfigHandler struct {
	db *gorm.DB
}

func NewConfigHandler(db *gorm.DB) *ConfigHandler {
	return &ConfigHandler{db: db}
}

type ConfigEntry struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

type configRecord struct {
	ID        uint      `gorm:"primaryKey"`
	Key       string    `gorm:"size:256;uniqueIndex"`
	Value     string    `gorm:"type:text"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (configRecord) TableName() string { return "configurations" }

type UpsertConfigRequest struct {
	Value string `json:"value" binding:"required"`
}

func (h *ConfigHandler) List(c *gin.Context) {
	var rows []configRecord
	if err := h.db.Order("key ASC").Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out := make([]ConfigEntry, 0, len(rows))
	for _, r := range rows {
		out = append(out, ConfigEntry{Key: r.Key, Value: r.Value, UpdatedAt: r.UpdatedAt})
	}
	c.JSON(http.StatusOK, out)
}

func (h *ConfigHandler) Get(c *gin.Context) {
	key := c.Param("key")
	var row configRecord
	err := h.db.Where("key = ?", key).First(&row).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ConfigEntry{Key: row.Key, Value: row.Value, UpdatedAt: row.UpdatedAt})
}

func (h *ConfigHandler) Set(c *gin.Context) {
	key := c.Param("key")
	var req UpsertConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	row := configRecord{
		Key:   key,
		Value: req.Value,
	}
	if err := h.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "updated_at"}),
	}).Create(&row).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Where("key = ?", key).First(&row).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ConfigEntry{Key: row.Key, Value: row.Value, UpdatedAt: row.UpdatedAt})
}

func (h *ConfigHandler) Delete(c *gin.Context) {
	key := c.Param("key")
	result := h.db.Where("key = ?", key).Delete(&configRecord{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *ConfigHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/configs", h.List)
	rg.GET("/configs/:key", h.Get)
	rg.PUT("/configs/:key", h.Set)
	rg.DELETE("/configs/:key", h.Delete)
}
