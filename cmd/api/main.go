package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/brainart16/brenox/internal/database"

	authHandler "github.com/brainart16/brenox/internal/auth"
	channelsHandler "github.com/brainart16/brenox/internal/channels"
	chatHandler "github.com/brainart16/brenox/internal/chat"
	middleware "github.com/brainart16/brenox/internal/middleware"
	realtimeHandler "github.com/brainart16/brenox/internal/realtime"

)

func main() {


	// Load environment variables from .env
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Failed to load environment variables")
	}

	
	// Create PostgreSQL connection pool
	pool, err := database.NewPostgresPool()

	if err != nil {
		log.Fatal(err)
	}


	// Create sqlc Queries object.
	// This becomes our database access layer.
	queries := db.New(pool)

	// Initialize real-time service (e.g., WebSocket hub).
	hub := realtimeHandler.NewHub()
	go hub.Run()

	// Initialize services and handlers.
	authService := authHandler.NewService(queries)
	authHandlerInstance := authHandler.NewHandler(authService)
	channelsService := channelsHandler.NewService(queries)
	channelsHandlerInstance := channelsHandler.NewHandler(channelsService)
	chatService := chatHandler.NewService(queries)
	chatHandlerInstance := chatHandler.NewHandler(chatService)

	wsHandler := realtimeHandler.NewHandler(hub, chatService, channelsService)


	// Gin router initialization
	router := gin.Default()

	// Register Auth routes
	authRouter := router.Group("/auth")
	authRouter.POST("/register", authHandlerInstance.Register)
	authRouter.POST("/login", authHandlerInstance.Login)

	// Register protected routes
	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware())

	// Register channel routes
	api.POST("/channels", channelsHandlerInstance.CreateChannel)
	api.GET("/channels", channelsHandlerInstance.GetChannels)
	api.POST("/channels/:id/messages", chatHandlerInstance.CreateMessage)
	api.GET("/channels/:id/messages", chatHandlerInstance.GetMessages)

	// Realtime routes
	api.GET("/ws", wsHandler.HandleWebSocket)
	api.GET("/presence", wsHandler.GetPresence)

	// Start HTTP server
	router.Run(":8080")
}