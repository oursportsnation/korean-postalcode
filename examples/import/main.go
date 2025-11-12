package main

import (
	"fmt"
	"log"
	"time"

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
	fmt.Println("🔌 Connecting to database...")
	db, err := gorm.Open(mysql.Open(cfg.Database.GetDSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate table
	fmt.Println("🔧 Creating table if not exists...")
	if err := db.AutoMigrate(&postalcode.PostalCodeRoad{}); err != nil {
		log.Fatalf("Failed to migrate: %v", err)
	}

	// Initialize service and importer
	repo := postalcodeapi.NewRepository(db)
	service := postalcodeapi.NewService(repo)
	importer := postalcodeapi.NewImporter(service)

	// Import data
	filePath := "../../docs/addresses/20251028_도로명범위.txt" // Adjust path as needed
	batchSize := cfg.Import.BatchSize

	fmt.Printf("📂 Importing from: %s\n", filePath)
	fmt.Printf("📦 Batch size: %d\n", batchSize)
	fmt.Println()

	startTime := time.Now()

	// Progress callback
	progressFn := func(current, total int) {
		percentage := float64(current) / float64(total) * 100
		fmt.Printf("✅ Progress: %d/%d (%.1f%%)\n", current, total, percentage)
	}

	// Execute import
	result, err := importer.ImportFromFile(filePath, batchSize, progressFn)
	if err != nil {
		log.Fatalf("❌ Import failed: %v", err)
	}

	duration := time.Since(startTime)

	// Print results
	fmt.Println()
	fmt.Println("📊 Import Summary:")
	fmt.Printf("  ✅ Success: %d records\n", result.TotalCount)
	fmt.Printf("  ❌ Errors:  %d records\n", result.ErrorCount)
	fmt.Printf("  ⏱️  Time:    %s\n", duration.Round(time.Second))

	if result.TotalCount > 0 {
		recordsPerSec := float64(result.TotalCount) / duration.Seconds()
		fmt.Printf("  📈 Speed:   %.0f records/sec\n", recordsPerSec)
	}

	fmt.Println()
	fmt.Println("🎉 Import completed successfully!")
}
