// @title Subscription Service API
// @version 1.0.0
// @description RESTful HTTP service for managing user subscriptions and cost aggregation

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginswagger "github.com/swaggo/gin-swagger"

	_ "sub-service/docs"
	"sub-service/internal/config"
	"sub-service/internal/handler"
	"sub-service/internal/middleware"
	"sub-service/internal/service"
	"sub-service/internal/store"
)

func main() {
	// 1. Load configuration

	cfg := config.Load()

	// 2. Setup structured logger

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// 3. Connect to database

	s, err := store.NewSubscriptionStore(cfg.Database)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	defer s.Close()

	// 4. Initialize app layers

	svc := service.NewSubscriptionService(s)
	h := handler.NewSubscriptionHandler(svc)

	// 5. Setup router with middleware

	r := gin.New()
	r.Use(
		middleware.RequestID(),
		middleware.Logger(logger),
		middleware.CORS([]string{"http://localhost:3000", "http://localhost:5173"}),
	)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))

	api := r.Group("/api/v1")
	h.RegisterRoutes(api)

	// 6. Configure HTTP server

	srv := &http.Server{
		Addr:    cfg.Server.Host + ":" + cfg.Server.Port,
		Handler: r,
	}

	// 7. Start server

	go func() {
		logger.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// 8. Wait for shutdown signal

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	// 9. Graceful shutdown

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	logger.Info("server exited")
}
