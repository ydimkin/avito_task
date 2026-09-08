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

	"core-service/internal/config"
	"core-service/internal/handler"
	clientHandler "core-service/internal/handler/client"
	partnerHandler "core-service/internal/handler/partner"
	applogger "core-service/internal/pkg/logger"
	"core-service/internal/repository"
	"core-service/internal/service"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("configuration load error", slog.Any("error", err))
		return
	}

	// Настройка логгирования через slog
	applogger.Setup(cfg.LogFormat, cfg.LogLevel)
	slog.Info("initializing core service",
		slog.String("port", cfg.Port),
		slog.String("log_level", cfg.LogLevel),
		slog.String("log_format", cfg.LogFormat),
	)

	// Инициализация подключения к базе данных PostgreSQL
	var db *gorm.DB

	db, err = gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})

	if err != nil {
		slog.Error("failed to connect to database", slog.Any("error", err))
		return
	}

	// Инициализация слоя репозиториев
	restaurantRepo := repository.NewPostgresRestaurantRepository(db)
	menuRepo := repository.NewPostgresMenuRepository(db)
	orderRepo := repository.NewPostgresOrderRepository(db)

	// Инициализация слоя сервисов
	webhookSvc := service.NewWebhookService(cfg.WebhookTimeoutSec)
	restaurantSvc := service.NewRestaurantService(restaurantRepo)
	menuSvc := service.NewMenuService(menuRepo, restaurantRepo)
	orderSvc := service.NewOrderService(orderRepo, menuRepo, restaurantRepo, webhookSvc)

	// Инициализация хендлеров
	routerCfg := handler.RouterConfig{
		RestaurantHandler: clientHandler.NewRestaurantHandler(restaurantSvc),
		MenuHandler:       clientHandler.NewMenuHandler(menuSvc),
		OrderHandler:      clientHandler.NewOrderHandler(orderSvc),
		PartnerHandler:    partnerHandler.NewHandler(menuSvc, orderSvc),
	}

	r := handler.SetupRouter(routerCfg)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("core service HTTP server listening", slog.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server fatal error", slog.Any("error", err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down core service HTTP server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("core service server forced to shutdown", slog.Any("error", err))
	} else {
		slog.Info("core service server exited cleanly")
	}
}
