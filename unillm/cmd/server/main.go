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
	apiadmin "github.com/unillm/unillm/api/admin"
	apidashboard "github.com/unillm/unillm/api/dashboard"
	apimiddleware "github.com/unillm/unillm/api/middleware"
	apiv1 "github.com/unillm/unillm/api/v1"
	corebilling "github.com/unillm/unillm/core/billing"
	corecatalog "github.com/unillm/unillm/core/catalog"
	coreinference "github.com/unillm/unillm/core/inference"
	infrabilling "github.com/unillm/unillm/infra/billing"
	infracatalog "github.com/unillm/unillm/infra/catalog"
	infracrypto "github.com/unillm/unillm/infra/crypto"
	infraInference "github.com/unillm/unillm/infra/inference"
	infrajwt "github.com/unillm/unillm/infra/jwt"
	"github.com/unillm/unillm/infra/persistence"
	"github.com/unillm/unillm/infra/provider"
	"github.com/unillm/unillm/internal/config"
	"github.com/unillm/unillm/internal/logger"
	"github.com/unillm/unillm/internal/service"
)

func main() {
	cfg := config.Load()

	logger.Init(cfg.Environment)

	if cfg.Environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := persistence.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect database")
	}
	if err := persistence.AutoMigrate(db); err != nil {
		log.Fatal().Err(err).Msg("failed to migrate database")
	}

	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse redis url")
	}
	rdb := redis.NewClient(opt)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal().Err(err).Msg("failed to connect redis")
	}

	userRepo := persistence.NewUserRepo(db)
	keyRepo := persistence.NewAPIKeyRepo(db)
	providerRepo := persistence.NewProviderRepo(db)
	usageRepo := persistence.NewUsageRepo(db)

	keyProtector := infracrypto.NewKeyProtector(cfg.EncryptionKey)

	billingStore := infrabilling.NewStore(rdb, db)
	billingSvc := corebilling.NewService(billingStore)

	catalogRepo := infracatalog.NewRepository(providerRepo, keyProtector)
	catalogSvc := corecatalog.NewService(catalogRepo)

	flushCtx, flushCancel := context.WithCancel(context.Background())
	go billingSvc.FlushWorker(flushCtx, 5*time.Second)

	registry := provider.NewRegistry()
	provider.RegisterAll(registry, providerRepo)

	if err := catalogSvc.Reload(context.Background()); err != nil {
		log.Warn().Err(err).Msg("failed to load provider keys")
	}

	inferenceSvc := coreinference.NewService(
		&infraInference.CatalogRoutes{Catalog: catalogSvc},
		&infraInference.ProviderRegistry{Registry: registry},
		&infraInference.BillingRecorder{Billing: billingSvc},
		&infraInference.MetricsAdapter{Record: apimiddleware.RecordProxy},
		infraInference.NewHTTPEmbeddingForwarder(30*time.Second),
	)

	jwtIssuer := infrajwt.NewIssuer(cfg.JWTSecret)
	authSvc := service.NewAuthService(userRepo, keyRepo, jwtIssuer, cfg.BcryptCost)

	authHandler := apidashboard.NewAuthHandler(authSvc)
	modelsHandler := apiv1.NewModelsHandler(providerRepo)
	proxyHandler := apiv1.NewProxyHandler(inferenceSvc)
	embeddingHandler := apiv1.NewEmbeddingHandler(inferenceSvc)
	usageHandler := apidashboard.NewUsageHandler(usageRepo, userRepo, billingSvc)
	adminHandler := apiadmin.NewHandler(providerRepo, userRepo, usageRepo, catalogSvc, keyProtector)
	statusHandler := apiv1.NewStatusHandler(registry, catalogSvc)

	statusHandler.StartHealthChecks(60 * time.Second)

	limiter := apimiddleware.NewRedisRateLimiter(rdb, 200, time.Minute)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(apimiddleware.RequestID())
	r.Use(apimiddleware.Metrics())
	r.Use(logger.GinLogger())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-Id"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-Id"},
		AllowCredentials: len(cfg.CORSOrigins) > 0 && cfg.CORSOrigins[0] != "*",
		MaxAge:           12 * time.Hour,
	}))

	r.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, cfg.MaxBodyBytes)
		c.Next()
	})

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "time": time.Now().UTC(), "env": cfg.Environment})
	})
	r.GET("/status", statusHandler.Status)
	r.GET("/status/history", statusHandler.StatusHistory)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.POST("/api/auth/register", authHandler.Register)
	r.POST("/api/auth/login", authHandler.Login)
	r.GET("/api/models/catalog", modelsHandler.ModelCatalog)

	dashboard := r.Group("/api")
	dashboard.Use(apimiddleware.JWTAuth(jwtIssuer))
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

	v1 := r.Group("/v1")
	v1.Use(apimiddleware.APIKeyAuth(authSvc.ResolveAPIKey))
	v1.Use(apimiddleware.RedisRateLimit(limiter))
	{
		v1.GET("/models", modelsHandler.ListModels)
		v1.POST("/chat/completions",
			apimiddleware.BalanceCheck(func(ctx context.Context, userID int64) (bool, error) {
				return billingSvc.CheckBalance(ctx, userID, 0)
			}),
			proxyHandler.ChatCompletion,
		)
		v1.POST("/embeddings",
			apimiddleware.BalanceCheck(func(ctx context.Context, userID int64) (bool, error) {
				return billingSvc.CheckBalance(ctx, userID, 0)
			}),
			embeddingHandler.CreateEmbedding,
		)
	}

	admin := r.Group("/api/admin")
	admin.Use(apimiddleware.JWTAuth(jwtIssuer))
	admin.Use(apimiddleware.AdminOnly())
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
