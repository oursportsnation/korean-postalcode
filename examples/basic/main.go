package main

import (
	"fmt"
	"log"

	postalcode "github.com/oursportsnation/korean-postalcode"
	postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// Load configuration from environment
	cfg, err := postalcode.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	db, err := gorm.Open(mysql.Open(cfg.Database.GetDSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize repository and service
	repo := postalcodeapi.NewRepository(db)
	service := postalcodeapi.NewService(repo)

	// Example 1: Search by exact zip code
	fmt.Println("=== Example 1: Search by Zip Code ===")
	results, err := service.GetByZipCode("01000")
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Found %d results for zip code 01000\n", len(results))
		if len(results) > 0 {
			fmt.Printf("First result: %s %s %s\n",
				results[0].SidoName,
				results[0].SigunguName,
				results[0].RoadName)
		}
	}

	// Example 2: Fast search by zip prefix (first 3 digits) with pagination
	fmt.Println("\n=== Example 2: Fast Search by Prefix (with pagination) ===")
	results, total, err := service.GetByZipPrefix("010", 10, 0)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Found %d results for prefix 010 (total: %d)\n", len(results), total)
	}

	// Example 3: Complex search with multiple filters
	fmt.Println("\n=== Example 3: Complex Search ===")
	params := postalcode.SearchParams{
		SidoName:    "서울특별시",
		SigunguName: "강북구",
		RoadName:    "삼양로",
		Page:        1,
		Limit:       10,
	}
	results2, total2, err := service.Search(params)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("Found %d results (total: %d)\n", len(results2), total2)
		for i, r := range results2 {
			fmt.Printf("  %d. [%s] %s %s %s (건물번호: %d~%d)\n",
				i+1,
				r.ZipCode,
				r.SidoName,
				r.SigunguName,
				r.RoadName,
				r.StartBuildingMain,
				*r.EndBuildingMain)
		}
	}
}
