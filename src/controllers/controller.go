package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func createRouter(db *gorm.DB, bot BotController) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(gin.Logger())

	v1 := engine.Group("/api/v1")

	configH := NewConfigHandler(db)
	configH.RegisterRoutes(v1)

	botH := NewBotHandler(bot)
	botH.RegisterRoutes(v1)

	return engine
}
