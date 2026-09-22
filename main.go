package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/ares/dp-vc-webApp/configs/env"
	"github.com/ares/dp-vc-webApp/configs/mongo"
	"github.com/ares/dp-vc-webApp/configs/observability"
	"github.com/ares/dp-vc-webApp/configs/redis"
	"github.com/ares/dp-vc-webApp/routes"
	"github.com/ares/dp-vc-webApp/worker"
)

var (
	isWorker             = flag.Bool("worker", false, "Run as background worker")
	runNotificationsOnce = flag.Bool("notifications-once", false, "Check and send payment notifications once")
)

func main() {
	flag.Parse()

	// Load environment configuration
	if err := env.Load(); err != nil {
		fmt.Printf("Warning: Error loading .env file: %v\n", err)
	}

	// Get configuration
	config := env.GetConfig()
	port := config.Port

	// Initialize OpenTelemetry (safe if no endpoint configured)
	tp, err := observability.Init(config)
	if err != nil {
		fmt.Printf("Warning: OpenTelemetry initialization error: %v\n", err)
	}
	defer observability.Shutdown(context.Background(), tp)

	// Initialize MongoDB
	_, err = mongo.Init(config)
	if err != nil {
		fmt.Printf("Error connecting to MongoDB: %v\n", err)
		os.Exit(1)
	}
	defer mongo.Disconnect(context.Background())

	// Ping MongoDB
	if err := mongo.Ping(context.Background()); err != nil {
		fmt.Printf("Warning: MongoDB ping failed: %v\n", err)
	} else {
		fmt.Println("MongoDB connection established")
	}

	// Initialize Redis
	_, err = redis.Init(config)
	if err != nil {
		fmt.Printf("Warning: Redis connection error: %v\n", err)
	} else {
		fmt.Println("Redis connection established")
	}
	defer redis.Close()

	// One-time notification check mode
	if *runNotificationsOnce {
		if err := worker.RunNotificationsOnce(context.Background()); err != nil {
			fmt.Printf("Notification scan failed: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Worker-only mode
	if *isWorker {
		runWorker()
		return
	}

	// Default: Run both HTTP server and worker (for Docker/production)
	runHTTPServerWithWorker(port)
}

func runWorker() {
	fmt.Println("Starting worker mode...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start worker with graceful shutdown
	worker.Start(ctx)
}

func runHTTPServerWithWorker(port string) {
	fmt.Println("Starting combined server and worker mode...")

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set Gin mode based on environment
	if env.GetConfig().AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()

	// Apply global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(CORSMiddleware())
	router.Use(ErrorHandlerMiddleware())
	router.Use(observability.Middleware())

	// Register routes
	routes.Register(router)

	// Create HTTP server
	srv := &http.Server{
		Addr:         "0.0.0.0:" + port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start worker in a goroutine
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		worker.Start(ctx)
	}()

	// Start server in goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("Server starting on http://0.0.0.0:%s\n", port)
		fmt.Printf("Health check: http://localhost:%s/health\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\nShutting down server and worker...")

	// Cancel context to stop worker
	cancel()

	// Graceful shutdown HTTP server with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		fmt.Printf("Server forced to shutdown: %v\n", err)
	}

	// Wait for worker and server to finish
	wg.Wait()

	fmt.Println("Server and worker stopped")
}

// CORSMiddleware handles CORS for all requests
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// ErrorHandlerMiddleware provides centralized error handling
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic
				fmt.Printf("Panic recovered: %v\n", r)

				// Get current span if available and record error
				observability.RecordError(c, fmt.Errorf("panic: %v", r))

				// Return JSON error response
				c.AbortWithStatusJSON(500, gin.H{
					"status": "error",
					"error":  "Internal server error",
				})
			}
		}()
		c.Next()
	}
}
