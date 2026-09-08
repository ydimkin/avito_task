package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"parent-service/internal/client"
	"parent-service/internal/config"
	"parent-service/internal/handler"
	"parent-service/internal/logger"
	"parent-service/internal/service"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("configuration load error", slog.Any("error", err))
		return
	}

	_ = logger.Setup(cfg.LogFormat, cfg.LogLevel)
	slog.Info("initializing parent/restaurant service",
		slog.String("port", cfg.Port),
		slog.String("core_service_url", cfg.CoreServiceURL),
		slog.String("restaurant_id", cfg.RestaurantID),
		slog.String("log_level", cfg.LogLevel),
	)

	coreClient := client.NewCoreClient(cfg.CoreServiceURL, cfg.RestaurantID)
	kitchenSim := service.NewKitchenSimulator(coreClient, true)
	restaurantHandler := handler.NewRestaurantHandler(kitchenSim)

	router := handler.SetupRouter(restaurantHandler)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("restaurant service HTTP server listening", slog.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("restaurant service fatal error", slog.Any("error", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down restaurant service HTTP server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("restaurant service server forced to shutdown", slog.Any("error", err))
	} else {
		slog.Info("restaurant service server exited cleanly")
	}
}
