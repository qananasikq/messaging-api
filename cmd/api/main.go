package main

import (
	"context"
	"errors"
	"expvar"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"messaging-api/internal/config"
	"messaging-api/internal/handlers"
	"messaging-api/internal/middleware"
	"messaging-api/internal/repositories"
	"messaging-api/internal/services"
	wshub "messaging-api/internal/websocket"
	jwtpkg "messaging-api/pkg/jwt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.MustLoad()

	// Настраиваем логгер один раз и сразу делаем его дефолтным
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.Log.Level(),
	}))
	slog.SetDefault(logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dbpool := mustInitPostgres(ctx, &cfg, logger)
	defer dbpool.Close()

	rdb := mustInitRedis(ctx, &cfg, logger)
	defer rdb.Close()

	jwt := jwtpkg.New(cfg.JWT.Secret, cfg.JWT.Issuer, cfg.JWT.AccessTTL)

	// Репозитории
	users := repositories.NewUserRepo(dbpool)
	dialogs := repositories.NewDialogRepo(dbpool)
	messages := repositories.NewMessageRepo(dbpool)

	// Кэш и сервисы
	cache := services.NewRedisCache(rdb, services.RedisCacheConfig{
		LastMessagesLimit: cfg.Cache.LastMessagesLimit,
		LastMessagesTTL:   cfg.Cache.LastMessagesTTL,
		UnreadTTL:         cfg.Cache.UnreadTTL,
	})

	userService := services.NewUserService(users, jwt)
	dialogService := services.NewDialogService(dialogs, messages, cache)
	messageService := services.NewMessageService(messages, dialogs, cache)

	// WebSocket hub в фоне
	hub := wshub.NewHub(logger)
	go hub.Run(ctx)

	// Собираем все зависимости для хендлеров
	handlerDeps := handlers.Deps{
		Logger:     logger,
		UserSvc:    userService,
		DialogSvc:  dialogService,
		MessageSvc: messageService,
		JWT:        jwt,
		WSHub:      hub,
		ReadyCheck: func(ctx context.Context) error {
			if err := dbpool.Ping(ctx); err != nil {
				return err
			}
			return rdb.Ping(ctx).Err()
		},
	}

	// Инициализация роутера
	router := setupRouter(&cfg, logger, handlerDeps)

	// Запуск сервера в горутине
	srv := &http.Server{
		Addr:              cfg.App.Addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("сервер запускается", "addr", cfg.App.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	// Ожидание сигнала на завершение
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var shutdownReason string

	select {
	case sig := <-sigCh:
		shutdownReason = fmt.Sprintf("получен сигнал %s", sig)
	case err := <-errCh:
		shutdownReason = fmt.Sprintf("ошибка сервера: %v", err)
		logger.Error(shutdownReason)
	}

	logger.Info("начинаем graceful shutdown", "причина", shutdownReason)

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// Закрываем все websocket-соединения
	hub.CloseAll(wshub.CloseReasonServerShutdown)

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("ошибка graceful shutdown http-сервера", "err", err)
	} else {
		logger.Info("http-сервер остановлен")
	}

	logger.Info("всё, до свидания ✌️")
}

func mustInitPostgres(ctx context.Context, cfg *config.Config, logger *slog.Logger) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, cfg.Postgres.DSN)
	if err != nil {
		logger.Error("не удалось подключиться к postgres", "err", err)
		os.Exit(1)
	}

	if err = pool.Ping(ctx); err != nil {
		logger.Error("ping postgres не прошёл", "err", err)
		os.Exit(1)
	}

	return pool
}

func mustInitRedis(ctx context.Context, cfg *config.Config, logger *slog.Logger) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		DialTimeout:  2 * time.Second,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Error("redis ping не прошёл", "err", err)
		os.Exit(1)
	}

	return client
}

func setupRouter(cfg *config.Config, logger *slog.Logger, deps handlers.Deps) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// CORS — только для локального фронта (потом можно будет расширить)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Общие middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.Recover(logger))
	r.Use(middleware.AccessLog(logger))
	r.Use(middleware.RateLimitPerUser(10, 20, time.Second))
	r.Use(middleware.SecurityHeaders())

	// Дебаг-эндпоинты (только если включены и с правильным токеном)
	if cfg.Debug.Enabled {
		debugGroup := r.Group(cfg.Debug.PathPrefix)
		debugGroup.Use(middleware.TokenAuth(cfg.Debug.Token))

		debugGroup.GET("/vars", gin.WrapH(expvar.Handler()))
		debugGroup.GET("/pprof/", gin.WrapH(http.HandlerFunc(pprof.Index)))
		debugGroup.GET("/pprof/cmdline", gin.WrapH(http.HandlerFunc(pprof.Cmdline)))
		debugGroup.GET("/pprof/profile", gin.WrapH(http.HandlerFunc(pprof.Profile)))
		debugGroup.GET("/pprof/symbol", gin.WrapH(http.HandlerFunc(pprof.Symbol)))
		debugGroup.GET("/pprof/trace", gin.WrapH(http.HandlerFunc(pprof.Trace)))
		debugGroup.GET("/pprof/allocs", gin.WrapH(pprof.Handler("allocs")))
		debugGroup.GET("/pprof/block", gin.WrapH(pprof.Handler("block")))
		debugGroup.GET("/pprof/goroutine", gin.WrapH(pprof.Handler("goroutine")))
		debugGroup.GET("/pprof/heap", gin.WrapH(pprof.Handler("heap")))
		debugGroup.GET("/pprof/mutex", gin.WrapH(pprof.Handler("mutex")))
		debugGroup.GET("/pprof/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
	}

	// Регистрируем все наши маршруты
	handlers.NewHandler(deps).RegisterRoutes(r)

	return r
}
