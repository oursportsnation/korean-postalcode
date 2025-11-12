package main

import (
	"flag"
	"fmt"
	"log"

	postalcode "github.com/oursportsnation/korean-postalcode"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 커맨드 라인 플래그
	dsn := flag.String("dsn", "", "MySQL DSN (optional: 없으면 .env 파일 사용)")
	command := flag.String("cmd", "up", "명령어: up (생성), down (삭제), fresh (재생성), status (상태 확인)")
	flag.Parse()

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
		fmt.Printf("✅ .env 파일에서 로드 완료 (DB: %s)\n", cfg.Database.Name)
	}

	validCommands := map[string]bool{
		"up":     true,
		"down":   true,
		"fresh":  true,
		"status": true,
	}

	if !validCommands[*command] {
		log.Fatal("\n❌ -cmd 는 'up', 'down', 'fresh', 'status' 중 하나여야 합니다")
	}

	fmt.Println("📦 Postal Code Migration Tool")
	fmt.Println("===================================")
	fmt.Printf("🔧 명령어: %s\n", *command)
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

	// 명령어 실행
	switch *command {
	case "up":
		runUp(db)
	case "down":
		runDown(db)
	case "fresh":
		runFresh(db)
	case "status":
		runStatus(db)
	}
}

// runUp은 테이블을 생성합니다.
func runUp(db *gorm.DB) {
	fmt.Println("🔼 테이블 생성 중...")
	fmt.Println()

	// 도로명주소 테이블
	fmt.Print("  📋 postal_code_roads 테이블... ")
	if err := db.AutoMigrate(&postalcode.PostalCodeRoad{}); err != nil {
		fmt.Println("❌")
		log.Fatalf("    에러: %v", err)
	}
	fmt.Println("✅")

	// 지번주소 테이블
	fmt.Print("  📋 postal_code_lands 테이블... ")
	if err := db.AutoMigrate(&postalcode.PostalCodeLand{}); err != nil {
		fmt.Println("❌")
		log.Fatalf("    에러: %v", err)
	}
	fmt.Println("✅")

	fmt.Println()
	fmt.Println("🎉 마이그레이션 완료!")
	fmt.Println()
	fmt.Println("💡 다음 단계:")
	fmt.Println("  1. 데이터 import: ./postalcode-import -dsn=\"...\" -file=\"data/postal_codes.txt\"")
	fmt.Println("  2. 또는 Shell 스크립트: ./scripts/import.sh")
	fmt.Println()
}

// runDown은 테이블을 삭제합니다.
func runDown(db *gorm.DB) {
	fmt.Println("🔽 테이블 삭제 중...")
	fmt.Println()

	// 지번주소 테이블 (외래키 고려하여 먼저 삭제)
	fmt.Print("  📋 postal_code_lands 테이블... ")
	if err := db.Migrator().DropTable(&postalcode.PostalCodeLand{}); err != nil {
		fmt.Println("❌")
		log.Fatalf("    에러: %v", err)
	}
	fmt.Println("✅")

	// 도로명주소 테이블
	fmt.Print("  📋 postal_code_roads 테이블... ")
	if err := db.Migrator().DropTable(&postalcode.PostalCodeRoad{}); err != nil {
		fmt.Println("❌")
		log.Fatalf("    에러: %v", err)
	}
	fmt.Println("✅")

	fmt.Println()
	fmt.Println("🎉 테이블 삭제 완료!")
	fmt.Println()
}

// runFresh는 테이블을 삭제하고 재생성합니다.
func runFresh(db *gorm.DB) {
	fmt.Println("🔄 테이블 재생성 중...")
	fmt.Println()

	runDown(db)
	fmt.Println("---")
	runUp(db)
}

// runStatus는 테이블 상태를 확인합니다.
func runStatus(db *gorm.DB) {
	fmt.Println("📊 테이블 상태 확인 중...")
	fmt.Println()

	// 도로명주소 테이블
	hasRoad := db.Migrator().HasTable(&postalcode.PostalCodeRoad{})
	fmt.Print("  📋 postal_code_roads: ")
	if hasRoad {
		fmt.Print("✅ 존재")
		var count int64
		db.Model(&postalcode.PostalCodeRoad{}).Count(&count)
		fmt.Printf(" (%d건)\n", count)
	} else {
		fmt.Println("❌ 없음")
	}

	// 지번주소 테이블
	hasLand := db.Migrator().HasTable(&postalcode.PostalCodeLand{})
	fmt.Print("  📋 postal_code_lands: ")
	if hasLand {
		fmt.Print("✅ 존재")
		var count int64
		db.Model(&postalcode.PostalCodeLand{}).Count(&count)
		fmt.Printf(" (%d건)\n", count)
	} else {
		fmt.Println("❌ 없음")
	}

	fmt.Println()

	if hasRoad && hasLand {
		fmt.Println("🎉 모든 테이블이 준비되었습니다!")
	} else {
		fmt.Println("⚠️  일부 테이블이 없습니다. 마이그레이션을 실행하세요:")
		fmt.Println("    ./postalcode-migrate -dsn=\"...\" -cmd=up")
	}
	fmt.Println()
}
