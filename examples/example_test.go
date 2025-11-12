package postalcode_test

import (
	"fmt"
	"net/http"

	postalcode "github.com/oursportsnation/korean-postalcode"
	postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
	"gorm.io/gorm"
)

// Example_basicUsage는 기본 사용법을 보여줍니다.
func Example_basicUsage() {
	var db *gorm.DB // your database connection

	// Repository와 Service 생성
	repo := postalcodeapi.NewRepository(db)
	service := postalcodeapi.NewService(repo)

	// 우편번호로 조회
	results, err := service.GetByZipCode("01000")
	if err != nil {
		panic(err)
	}

	for _, road := range results {
		fmt.Printf("%s %s %s\n", road.SidoName, road.SigunguName, road.RoadName)
	}

	// 우편번호 앞 3자리로 빠른 조회 (페이지네이션 적용)
	results, total, err := service.GetByZipPrefix("010", 10, 0)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d roads (total: %d)\n", len(results), total)
}

// Example_search는 복합 검색 사용법을 보여줍니다.
func Example_search() {
	var db *gorm.DB // your database connection

	repo := postalcodeapi.NewRepository(db)
	service := postalcodeapi.NewService(repo)

	// 복합 검색
	params := postalcode.SearchParams{
		SidoName:    "서울",
		SigunguName: "강북구",
		RoadName:    "삼양로",
		Page:        1,
		Limit:       10,
	}

	results, total, err := service.Search(params)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Found %d of %d total results\n", len(results), total)
}

// Example_restAPI는 REST API 사용법을 보여줍니다.
func Example_restAPI() {
	var db *gorm.DB // your database connection

	// Setup
	repo := postalcodeapi.NewRepository(db)
	service := postalcodeapi.NewService(repo)

	// 표준 http.ServeMux 사용
	mux := http.NewServeMux()
	postalcodeapi.RegisterHTTPRoutes(service, mux, "/api/v1/postal-codes")

	// 서버 시작
	http.ListenAndServe(":8080", mux)
}

// Example_upsert는 데이터 생성/업데이트 사용법을 보여줍니다.
func Example_upsert() {
	var db *gorm.DB // your database connection

	repo := postalcodeapi.NewRepository(db)
	service := postalcodeapi.NewService(repo)

	// 단일 데이터 추가/업데이트
	road := &postalcode.PostalCodeRoad{
		ZipCode:           "01000",
		SidoName:          "서울특별시",
		SidoNameEn:        "Seoul",
		SigunguName:       "강북구",
		SigunguNameEn:     "Gangbuk-gu",
		RoadName:          "삼양로",
		RoadNameEn:        "Samyang-ro",
		StartBuildingMain: 689,
		RangeType:         1,
	}

	err := service.Upsert(road)
	if err != nil {
		panic(err)
	}

	fmt.Println("Upserted successfully")
}

// Example_batchUpsert는 배치 import 사용법을 보여줍니다.
func Example_batchUpsert() {
	var db *gorm.DB // your database connection

	repo := postalcodeapi.NewRepository(db)
	service := postalcodeapi.NewService(repo)

	// 배치 데이터 준비
	roads := []postalcode.PostalCodeRoad{
		{
			ZipCode:           "01000",
			SidoName:          "서울특별시",
			SigunguName:       "강북구",
			RoadName:          "삼양로177길",
			StartBuildingMain: 93,
			RangeType:         3,
		},
		{
			ZipCode:           "01000",
			SidoName:          "서울특별시",
			SigunguName:       "강북구",
			RoadName:          "삼양로179길",
			StartBuildingMain: 12,
			RangeType:         3,
		},
	}

	err := service.BatchUpsert(roads)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Batch upserted %d roads\n", len(roads))
}
