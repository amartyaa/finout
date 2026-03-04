package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"

	"github.com/finops/backend/internal/config"
	"github.com/finops/backend/internal/database"
	"github.com/finops/backend/internal/handlers"
	"github.com/finops/backend/internal/middleware"
	"github.com/finops/backend/internal/workers"
)

func main() {
	// Load .env file (ignore error, env vars may be set directly)
	godotenv.Load()

	cfg := config.Load()

	// Connect to PostgreSQL
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	_, filename, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(filename), "..", "..", "internal", "database", "migrations")
	if err := database.RunMigrations(db, migrationsDir); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Connect to Redis
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}
	rdb := redis.NewClient(opt)
	defer rdb.Close()

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("✅ Connected to Redis")

	// Initialize handlers
	authHandler := &handlers.AuthHandler{DB: db, JWTSecret: cfg.JWTSecret}
	orgHandler := &handlers.OrgHandler{DB: db}
	awsHandler := &handlers.AWSHandler{DB: db, Redis: rdb, Config: cfg}
	insightsHandler := &handlers.InsightsHandler{DB: db}

	// Start background workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	costWorker := workers.NewCostSyncWorker(db, rdb, cfg)
	go costWorker.Start(ctx)

	// Setup Gin router
	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:3001"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "finops-api"})
	})

	// Auth routes (public)
	auth := r.Group("/api/auth")
	{
		auth.POST("/signup", authHandler.Signup)
		auth.POST("/login", authHandler.Login)
	}

	// Protected routes
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(cfg.JWTSecret))
	{
		api.GET("/auth/me", authHandler.Me)

		// Organization routes
		api.POST("/orgs", orgHandler.CreateOrg)
		api.GET("/orgs", orgHandler.ListOrgs)

		// Org-scoped routes
		org := api.Group("/orgs/:orgId")
		org.Use(middleware.TenantMiddleware(db))
		{
			// Projects
			org.POST("/projects", orgHandler.CreateProject)
			org.GET("/projects", orgHandler.ListProjects)
			org.POST("/projects/:projectId/environments", orgHandler.CreateEnvironment)

			// AWS Connection
			org.POST("/aws/connect", awsHandler.Connect)
			org.GET("/aws/status", awsHandler.Status)
			org.POST("/aws/sync", awsHandler.TriggerSync)

			// Insights
			org.GET("/insights/overview", insightsHandler.Overview)
			org.GET("/insights/anomalies", insightsHandler.ListAnomalies)
			org.GET("/insights/forecast", insightsHandler.GetForecast)
			org.GET("/insights/recommendations", insightsHandler.ListRecommendations)
			org.POST("/insights/recommendations/:recId/dismiss", insightsHandler.DismissRecommendation)
		}
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		cancel()
	}()

	port := cfg.Port
	log.Printf("🚀 FinOps API server starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
