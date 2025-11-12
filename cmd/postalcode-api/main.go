package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	postalcode "github.com/oursportsnation/korean-postalcode"
	postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	_ "github.com/oursportsnation/korean-postalcode/docs/swagger" // Swagger docs
)

var (
	port   = flag.String("port", "8080", "Server port")
	host   = flag.String("host", "0.0.0.0", "Server host")
	dsn    = flag.String("dsn", "", "Database DSN (overrides .env)")
	envDir = flag.String("env", ".", "Directory containing .env file")
)

// @title Korean PostalCode API
// @version 1.0
// @description 한국 우편번호 검색 API - 도로명주소 및 지번주소 지원
// @description 행정안전부 우편번호 데이터를 기반으로 한 우편번호 검색 서비스입니다.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/oursportsnation/korean-postalcode
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @schemes http https
func main() {
	flag.Parse()

	// Load configuration
	var cfg *postalcode.Config
	var err error

	// Change to env directory if specified
	if *envDir != "." {
		if err := os.Chdir(*envDir); err != nil {
			log.Printf("Warning: Failed to change to env directory %s: %v", *envDir, err)
		}
	}

	cfg, err = postalcode.LoadConfig()
	if err != nil {
		log.Printf("Warning: Failed to load config from .env: %v", err)
		log.Println("Using default configuration or command-line arguments")
		cfg = &postalcode.Config{}
	}

	// Determine DSN to use
	var dbDSN string
	if *dsn != "" {
		// Use command-line DSN if provided
		dbDSN = *dsn
	} else {
		// Use config DSN
		dbDSN = cfg.Database.GetDSN()
	}

	// Validate DSN
	if dbDSN == "" {
		log.Fatal("❌ Database DSN is required. Use -dsn flag or set in .env file")
	}

	// Connect to database
	log.Println("📦 Connecting to database...")
	db, err := gorm.Open(mysql.Open(dbDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	log.Println("✅ Database connected successfully")

	// Auto migrate tables
	log.Println("🔧 Running auto migrations...")
	if err := db.AutoMigrate(&postalcode.PostalCodeRoad{}, &postalcode.PostalCodeLand{}); err != nil {
		log.Fatalf("❌ Failed to migrate database: %v", err)
	}
	log.Println("✅ Migrations completed")

	// Initialize service
	repo := postalcodeapi.NewRepository(db)
	service := postalcodeapi.NewService(repo)

	// Setup Gin router
	gin.SetMode(gin.ReleaseMode) // Use gin.DebugMode for development
	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "korean-postalcode",
			"version": "1.0.0",
		})
	})

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		postalCodes := v1.Group("/postal-codes")
		postalcodeapi.RegisterGinRoutes(service, postalCodes)
	}

	// Print startup information
	addr := fmt.Sprintf("%s:%s", *host, *port)
	printStartupInfo(addr)

	// Setup HTTP server with graceful shutdown
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Server starting on http://%s\n", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("\n🛑 Shutting down server...")

	// Graceful shutdown with 5 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server exited gracefully")
}

// corsMiddleware adds CORS headers
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// printStartupInfo prints server startup information
func printStartupInfo(addr string) {
	fmt.Println("\n" + "═══════════════════════════════════════════════════════════════")
	fmt.Println("🇰🇷  Korean PostalCode API Server")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("📡 Server Address:")
	fmt.Printf("   http://%s\n", addr)
	fmt.Println()
	fmt.Println("📍 도로명주소 API Endpoints:")
	fmt.Printf("   GET  http://%s/api/v1/postal-codes/road/zipcode/{code}\n", addr)
	fmt.Printf("   GET  http://%s/api/v1/postal-codes/road/prefix/{prefix}\n", addr)
	fmt.Printf("   GET  http://%s/api/v1/postal-codes/road/search\n", addr)
	fmt.Println()
	fmt.Println("📍 지번주소 API Endpoints:")
	fmt.Printf("   GET  http://%s/api/v1/postal-codes/land/zipcode/{code}\n", addr)
	fmt.Printf("   GET  http://%s/api/v1/postal-codes/land/prefix/{prefix}\n", addr)
	fmt.Printf("   GET  http://%s/api/v1/postal-codes/land/search\n", addr)
	fmt.Println()
	fmt.Println("🔍 도로명주소 Example Requests:")
	fmt.Printf("   curl http://%s/api/v1/postal-codes/road/zipcode/01000\n", addr)
	fmt.Printf("   curl http://%s/api/v1/postal-codes/road/prefix/010\n", addr)
	fmt.Printf("   curl 'http://%s/api/v1/postal-codes/road/search?sido_name=서울&limit=10'\n", addr)
	fmt.Println()
	fmt.Println("🔍 지번주소 Example Requests:")
	fmt.Printf("   curl http://%s/api/v1/postal-codes/land/zipcode/25627\n", addr)
	fmt.Printf("   curl http://%s/api/v1/postal-codes/land/prefix/256\n", addr)
	fmt.Printf("   curl 'http://%s/api/v1/postal-codes/land/search?sido_name=강원&eupmyeondong_name=강동면'\n", addr)
	fmt.Println()
	fmt.Println("🏥 Health Check:")
	fmt.Printf("   curl http://%s/health\n", addr)
	fmt.Println()
	fmt.Println("📚 Swagger API Documentation:")
	fmt.Printf("   http://%s/swagger/index.html\n", addr)
	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()
}
