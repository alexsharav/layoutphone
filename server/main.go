package main

import (
	"github.com/alexsharav/layout-phone-backend/internal/config"
	"github.com/alexsharav/layout-phone-backend/internal/handlers"
	"github.com/alexsharav/layout-phone-backend/internal/middleware"
	"github.com/alexsharav/layout-phone-backend/internal/services"
	"github.com/alexsharav/layout-phone-backend/internal/store"
	"github.com/alexsharav/layout-phone-backend/internal/websocket"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	memoryStore := store.NewMemoryStore()
	hub := websocket.NewHub()
	gameService := services.NewGameService(memoryStore, hub, cfg.TopicDefaultTime, cfg.RoundDefaultTime)

	router := gin.Default()
	router.Use(middleware.CORS(cfg.FrontendURL))

	handlers.RegisterRoutes(router, gameService, hub)

	if err := router.Run(":" + cfg.Port); err != nil {
		panic(err)
	}
}
