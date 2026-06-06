package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	db "github.com/brainart16/brenox/internal/db"
	"github.com/brainart16/brenox/internal/database"
	"github.com/brainart16/brenox/internal/health"
	redisutil "github.com/brainart16/brenox/internal/redis"

	authHandler "github.com/brainart16/brenox/internal/auth"
	"github.com/brainart16/brenox/internal/authz"
	channelsHandler "github.com/brainart16/brenox/internal/channels"
	chatHandler "github.com/brainart16/brenox/internal/chat"
	middleware "github.com/brainart16/brenox/internal/middleware"
	realtimeHandler "github.com/brainart16/brenox/internal/realtime"
	workspacesHandler "github.com/brainart16/brenox/internal/workspaces"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load environment variables")
	}

	pool, err := database.NewPostgresPool()
	if err != nil {
		log.Fatal(err)
	}

	redisClient, err := redisutil.NewClient()
	if err != nil {
		slog.Warn("redis unavailable, running in local-only realtime mode", "error", err)
	}

	queries := db.New(pool)
	authzService := authz.NewService(queries)
	wsConfig := realtimeHandler.LoadConfig()

	hub := realtimeHandler.NewHub(wsConfig)
	broker := realtimeHandler.NewBroker(redisClient, hub)
	hub.SetBroker(broker)
	go hub.Run()
	broker.Start()

	authService := authHandler.NewService(queries)
	authHandlerInstance := authHandler.NewHandler(authService)

	workspacesService := workspacesHandler.NewService(queries, authzService)
	workspacesHandlerInstance := workspacesHandler.NewHandler(workspacesService)

	channelsService := channelsHandler.NewService(queries, authzService)
	channelsHandlerInstance := channelsHandler.NewHandler(channelsService, hub)

	chatService := chatHandler.NewService(queries, authzService)
	chatHandlerInstance := chatHandler.NewHandler(chatService)

	wsHandler := realtimeHandler.NewHandler(hub, chatService, channelsService, wsConfig)
	healthHandler := health.NewHandler(pool, redisClient)

	router := gin.Default()

	router.GET("/health", healthHandler.Check)

	authRouter := router.Group("/auth")
	authRouter.POST("/register", authHandlerInstance.Register)
	authRouter.POST("/login", authHandlerInstance.Login)

	api := router.Group("/api")
	api.Use(middleware.AuthMiddleware())

	api.POST("/workspaces", workspacesHandlerInstance.CreateWorkspace)
	api.GET("/workspaces", workspacesHandlerInstance.ListWorkspaces)
	api.GET("/workspaces/:workspace_id", workspacesHandlerInstance.GetWorkspace)

	workspaceAPI := api.Group("/workspaces/:workspace_id")
	workspaceAPI.GET("/members", workspacesHandlerInstance.ListMembers)
	workspaceAPI.POST("/members", workspacesHandlerInstance.AddMember)
	workspaceAPI.DELETE("/members/:user_id", workspacesHandlerInstance.RemoveMember)
	workspaceAPI.PATCH("/members/:user_id", workspacesHandlerInstance.UpdateMemberRole)
	workspaceAPI.POST("/channels", channelsHandlerInstance.CreateChannel)
	workspaceAPI.GET("/channels", channelsHandlerInstance.GetChannels)
	workspaceAPI.POST("/channels/:id/join", channelsHandlerInstance.JoinChannel)
	workspaceAPI.POST("/channels/:id/leave", channelsHandlerInstance.LeaveChannel)
	workspaceAPI.POST("/channels/:id/messages", chatHandlerInstance.CreateMessage)
	workspaceAPI.GET("/channels/:id/messages", chatHandlerInstance.GetMessages)

	api.GET("/ws", wsHandler.HandleWebSocket)
	api.GET("/presence", wsHandler.GetPresence)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		slog.Info("http server listening", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	broker.Close()
	hub.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("server shutdown failed", "error", err)
	}
}
