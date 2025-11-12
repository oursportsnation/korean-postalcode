package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	postalcode "github.com/oursportsnation/korean-postalcode"
	postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 커맨드 라인 플래그
	dsn := flag.String("dsn", "", "MySQL DSN (optional: 없으면 .env 파일 사용)")
	filePath := flag.String("file", "", "주소 데이터 파일 경로 (required)")
	dataType := flag.String("type", "road", "데이터 타입: road (도로명주소) 또는 land (지번주소)")
	batchSize := flag.Int("batch", 1000, "배치 처리 사이즈")
	flag.Parse()

	if *filePath == "" {
		flag.Usage()
		log.Fatal("\n❌ -file 은 필수입니다")
	}

	// DSN 결정: 플래그 우선, 없으면 .env 파일
	var finalDSN string
	if *dsn != "" {
		finalDSN = *dsn
	} else {
		// .env 파일에서 설정 로드
		fmt.Println("📄 .env 파일에서 설정 로드 중...")
		cfg, err := postalcode.LoadConfig()
		if err != nil {
			log.Fatal("\n❌ .env 파일 로드 실패 및 -dsn 플래그 없음\n💡 해결방법:\n  1. -dsn 플래그 사용: -dsn=\"user:pass@tcp(host:port)/dbname\"\n  2. .env 파일 생성 (configs/.env.example 참고)")
		}
		finalDSN = cfg.Database.GetDSN()
		fmt.Printf("✅ .env 파일에서 로드 완료 (DB: %s)\n\n", cfg.Database.Name)
	}

	if *dataType != "road" && *dataType != "land" {
		log.Fatal("\n❌ -type 은 'road' 또는 'land' 여야 합니다")
	}

	typeKorean := "도로명주소"
	if *dataType == "land" {
		typeKorean = "지번주소"
	}

	fmt.Println("📍 Postal Code Import Tool")
	fmt.Println("===================================")
	fmt.Printf("📂 파일: %s\n", *filePath)
	fmt.Printf("📋 타입: %s (%s)\n", *dataType, typeKorean)
	fmt.Printf("📦 배치 사이즈: %d\n", *batchSize)
	fmt.Println()

	// 데이터베이스 연결
	fmt.Println("🔌 데이터베이스 연결 중...")
	db, err := gorm.Open(mysql.Open(finalDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ 데이터베이스 연결 실패: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("❌ DB 인스턴스 가져오기 실패: %v", err)
	}
	defer sqlDB.Close()

	fmt.Println("✅ 데이터베이스 연결 성공")
	fmt.Println()

	// 테이블 자동 생성 (필요한 경우)
	fmt.Println("🔧 테이블 확인 중...")
	if *dataType == "road" {
		if err := db.AutoMigrate(&postalcode.PostalCodeRoad{}); err != nil {
			log.Fatalf("❌ 테이블 생성 실패: %v", err)
		}
	} else {
		if err := db.AutoMigrate(&postalcode.PostalCodeLand{}); err != nil {
			log.Fatalf("❌ 테이블 생성 실패: %v", err)
		}
	}
	fmt.Println("✅ 테이블 준비 완료")
	fmt.Println()

	// PostalCode Service & Importer 생성
	repo := postalcodeapi.NewRepository(db)
	service := postalcodeapi.NewService(repo)
	importer := postalcodeapi.NewImporter(service)

	// Import 시작
	fmt.Println("🔄 데이터 가져오기 시작...")
	startTime := time.Now()

	// 진행 상황 콜백
	progressFn := func(current, total int) {
		fmt.Printf("✅ 처리됨: %d / %d건 (%.1f%%)\n", current, total, float64(current)/float64(total)*100)
	}

	// Import 실행
	var result *postalcode.ImportResult

	var importErr error
	if *dataType == "road" {
		fmt.Println("📍 도로명주소 데이터 import 중...")
		result, importErr = importer.ImportFromFile(*filePath, *batchSize, progressFn)
	} else {
		fmt.Println("📍 지번주소 데이터 import 중...")
		result, importErr = importer.ImportLandFromFile(*filePath, *batchSize, progressFn)
	}

	if importErr != nil {
		log.Fatalf("❌ Import 실패: %v", importErr)
	}

	duration := time.Since(startTime)

	fmt.Println()
	fmt.Printf("📊 Import 완료!\n")
	fmt.Printf("  - 타입: %s\n", typeKorean)
	fmt.Printf("  - 성공: %d건\n", result.TotalCount)
	fmt.Printf("  - 실패: %d건\n", result.ErrorCount)
	fmt.Printf("  - 소요 시간: %s\n", duration.Round(time.Second))
	fmt.Println()
}
