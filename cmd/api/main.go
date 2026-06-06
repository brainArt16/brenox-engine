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
	workspacesHandler "github.com/brainart16/brenox/internal/workspaces"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load environment variables")
	}

	pool, err := database.NewPostgresPool()
	if err != nil {
		log.Fatal(err)
	}

	queries := db.New(pool)

	hub := realtimeHandler.NewHub()
	go hub.Run()

	authService := authHandler.NewService(queries)
	authHandlerInstance := authHandler.NewHandler(authService)

	workspacesService := workspacesHandler.NewService(queries)
	workspacesHandlerInstance := workspacesHandler.NewHandler(workspacesService)

	channelsService := channelsHandler.NewService(queries)
	channelsHandlerInstance := channelsHandler.NewHandler(channelsService, hub)

	chatService := chatHandler.NewService(queries)
	chatHandlerInstance := chatHandler.NewHandler(chatService)

	wsHandler := realtimeHandler.NewHandler(hub, chatService, channelsService)

	router := gin.Default()

	authRouter := router.Group("/auth")
	authRouter.POST("/register", authHandlerInstance.Register)
	authRouter.POST("/login", authHandlerInstance.Login)

	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware())

	api.POST("/workspaces", workspacesHandlerInstance.CreateWorkspace)
	api.GET("/workspaces", workspacesHandlerInstance.ListWorkspaces)
	api.GET("/workspaces/:workspace_id", workspacesHandlerInstance.GetWorkspace)

	workspaceAPI := api.Group("/workspaces/:workspace_id")
	workspaceAPI.POST("/channels", channelsHandlerInstance.CreateChannel)
	workspaceAPI.GET("/channels", channelsHandlerInstance.GetChannels)
	workspaceAPI.POST("/channels/:id/join", channelsHandlerInstance.JoinChannel)
	workspaceAPI.POST("/channels/:id/leave", channelsHandlerInstance.LeaveChannel)
	workspaceAPI.POST("/channels/:id/messages", chatHandlerInstance.CreateMessage)
	workspaceAPI.GET("/channels/:id/messages", chatHandlerInstance.GetMessages)

	api.GET("/ws", wsHandler.HandleWebSocket)
	api.GET("/presence", wsHandler.GetPresence)

	router.Run(":8080")
}
