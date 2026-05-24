package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/unillm/unillm/internal/config"
	"github.com/unillm/unillm/internal/handler"
	"github.com/unillm/unillm/internal/logger"
	"github.com/unillm/unillm/internal/middleware"
	"github.com/unillm/unillm/internal/provider"
	"github.com/unillm/unillm/internal/repository"
	"github.com/unillm/unillm/internal/service"
)

func main() {
	cfg := config.Load()

	// Initialize structured logging
	logger.Init(cfg.Environment)

	if cfg.Environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Database
	db, err := repository.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect database")
	}
	if err := repository.AutoMigrate(db); err != nil {
		log.Fatal().Err(err).Msg("failed to migrate database")
	}

	// Redis
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse redis url")
	}
	rdb := redis.NewClient(opt)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal().Err(err).Msg("failed to connect redis")
	}

	// Repositories
	userRepo := repository.NewUserRepo(db)
	keyRepo := repository.NewAPIKeyRepo(db)
	providerRepo := repository.NewProviderRepo(db)

	// Services
	authSvc := service.NewAuthService(userRepo, keyRepo, cfg.JWTSecret, cfg.BcryptCost)
	billingSvc := service.NewBillingService(rdb, db)

	// Start billing flush worker (Redis → PG every 5 seconds)
	flushCtx, flushCancel := context.WithCancel(context.Background())
	go billingSvc.FlushWorker(flushCtx, 5*time.Second)

	// Provider registry with resilience (retry + circuit breaker)
	registry := provider.NewRegistry()
	registerProviders(registry, providerRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	modelsHandler := handler.NewModelsHandler(providerRepo, db)
	proxyHandler := handler.NewProxyHandler(registry, providerRepo, billingSvc)
	embeddingHandler := handler.NewEmbeddingHandler(providerRepo, billingSvc)
	usageHandler := handler.NewUsageHandler(db, billingSvc)
	adminHandler := handler.NewAdminHandler(db, providerRepo, userRepo)
	statusHandler := handler.NewStatusHandler(registry, providerRepo)

	// Wire status handler to proxy for active probing
	statusHandler.SetProxyHandler(proxyHandler)

	// Start background health checks (every 60 seconds)
	statusHandler.StartHealthChecks(60 * time.Second)

	// Load upstream API keys
	if err := proxyHandler.LoadProviderKeys(providerRepo); err != nil {
		log.Warn().Err(err).Msg("failed to load provider keys")
	}
	if err := embeddingHandler.LoadProviderKeys(providerRepo); err != nil {
		log.Warn().Err(err).Msg("failed to load embedding provider keys")
	}

	// Rate limiter: 200 requests per minute per user
	limiter := middleware.NewRateLimiter(200, time.Minute)

	// Router (use New instead of Default to avoid double logging)
	r := gin.New()
	r.Use(gin.Recovery())

	// Global middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.Metrics())
	r.Use(logger.GinLogger())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-Id"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-Id"},
		AllowCredentials: len(cfg.CORSOrigins) > 0 && cfg.CORSOrigins[0] != "*",
		MaxAge:           12 * time.Hour,
	}))

	// Request body size limit
	r.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, cfg.MaxBodyBytes)
		c.Next()
	})

	// Health & status (public, no auth)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "time": time.Now().UTC(), "env": cfg.Environment})
	})
	r.GET("/status", statusHandler.Status)
	r.GET("/status/history", statusHandler.StatusHistory)

	// Prometheus metrics
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Public routes (no auth)
	r.POST("/api/auth/register", authHandler.Register)
	r.POST("/api/auth/login", authHandler.Login)
	r.GET("/api/models/catalog", modelsHandler.ModelCatalog)

	// Dashboard API (JWT auth)
	dashboard := r.Group("/api")
	dashboard.Use(middleware.JWTAuth(cfg.JWTSecret))
	{
		dashboard.GET("/me", authHandler.Me)
		dashboard.POST("/keys", authHandler.CreateAPIKey)
		dashboard.GET("/keys", authHandler.ListAPIKeys)
		dashboard.DELETE("/keys", authHandler.DeleteAPIKey)
		dashboard.PUT("/password", authHandler.ChangePassword)
		dashboard.GET("/usage/summary", usageHandler.Summary)
		dashboard.GET("/usage/by-model", usageHandler.ByModel)
		dashboard.GET("/usage/daily", usageHandler.Daily)
		dashboard.GET("/usage/recent", usageHandler.Recent)
	}

	// OpenAI-compatible proxy (API key auth + balance check)
	v1 := r.Group("/v1")
	v1.Use(middleware.APIKeyAuth(authSvc.ResolveAPIKey))
	v1.Use(middleware.RateLimit(limiter))
	{
		v1.GET("/models", modelsHandler.ListModels)
		v1.POST("/chat/completions",
			middleware.BalanceCheck(func(ctx context.Context, userID int64) (bool, error) {
				return billingSvc.CheckBalance(ctx, userID, 0)
			}),
			proxyHandler.ChatCompletion,
		)
		v1.POST("/embeddings",
			middleware.BalanceCheck(func(ctx context.Context, userID int64) (bool, error) {
				return billingSvc.CheckBalance(ctx, userID, 0)
			}),
			embeddingHandler.CreateEmbedding,
		)
	}

	// Admin API (JWT auth + admin role)
	admin := r.Group("/api/admin")
	admin.Use(middleware.JWTAuth(cfg.JWTSecret))
	admin.Use(middleware.AdminOnly())
	{
		admin.GET("/stats", adminHandler.GlobalStats)
		admin.GET("/users", adminHandler.ListUsers)
		admin.POST("/users/balance", adminHandler.UpdateUserBalance)
		admin.GET("/providers", adminHandler.ListProviders)
		admin.POST("/providers", adminHandler.CreateProvider)
		admin.PUT("/providers/toggle", adminHandler.ToggleProvider)
		admin.GET("/models", adminHandler.ListModels)
		admin.POST("/models", adminHandler.CreateModel)
		admin.PUT("/models", adminHandler.UpdateModel)
		admin.GET("/provider-keys", adminHandler.ListProviderKeys)
		admin.POST("/provider-keys", adminHandler.AddProviderKey)
	}

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		log.Info().Str("port", cfg.Port).Str("env", cfg.Environment).Msg("UniLLM server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("server forced to shutdown")
	}

	flushCancel()
	log.Info().Msg("flushing remaining billing data...")
	billingSvc.FlushAll(context.Background())

	log.Info().Msg("server stopped")
}

func registerProviders(registry *provider.Registry, providerRepo *repository.ProviderRepo) {
	providers, err := providerRepo.ListActive()
	if err != nil {
		log.Warn().Err(err).Msg("failed to load providers")
		return
	}
	for _, p := range providers {
		var inner provider.Provider
		switch p.Name {
		case "openai", "deepseek", "alibaba", "bytedance", "geneasy":
			inner = provider.NewOpenAIProvider(p.Name, p.BaseURL)
		case "anthropic":
			inner = provider.NewAnthropicProvider(p.BaseURL)
		case "google":
			inner = provider.NewGoogleProvider(p.BaseURL)
		default:
			log.Warn().Str("provider", p.Name).Msg("unknown provider")
			continue
		}
		// Wrap with retry + circuit breaker
		registry.Register(provider.NewResilientProvider(inner))
		log.Info().Str("provider", p.Name).Str("url", p.BaseURL).Msg("registered provider")
	}
}
