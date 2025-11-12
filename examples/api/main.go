package main

import (
	"fmt"
	"log"
	"net/http"

	postalcode "github.com/oursportsnation/korean-postalcode"
	postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg, err := postalcode.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := gorm.Open(mysql.Open(cfg.Database.GetDSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize service and register routes
	repo := postalcodeapi.NewRepository(db)
	service := postalcodeapi.NewService(repo)

	// Setup HTTP router
	mux := http.NewServeMux()

	// Register PostalCode API routes
	postalcodeapi.RegisterHTTPRoutes(service, mux, "/api/v1/postal-codes")

	// Add health check endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok"}`)
	})

	// Start server
	addr := ":8080"
	fmt.Printf("🚀 PostalCode API Server starting on %s\n", addr)
	fmt.Printf("📍 API endpoints:\n")
	fmt.Printf("   GET  %s/zipcode/{code}     - Search by zip code\n", "/api/v1/postal-codes")
	fmt.Printf("   GET  %s/prefix/{prefix}    - Fast search by prefix\n", "/api/v1/postal-codes")
	fmt.Printf("   GET  %s/search             - Complex search\n", "/api/v1/postal-codes")
	fmt.Printf("   GET  /health                        - Health check\n")
	fmt.Printf("\n")
	fmt.Printf("🔍 Example requests:\n")
	fmt.Printf("   curl http://localhost:8080/api/v1/postal-codes/{road|land}/zipcode/01000\n")
	fmt.Printf("   curl http://localhost:8080/api/v1/postal-codes/{road|land}/prefix/010\n")
	fmt.Printf("   curl 'http://localhost:8080/api/v1/postal-codes/{road|land}/search?sido_name=서울&limit=10'\n")
	fmt.Printf("\n")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
