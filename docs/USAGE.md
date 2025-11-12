# PostalCode 패키지 완전 가이드

## 목차
1. [기본 사용법](#기본-사용법)
2. [데이터 Import](#데이터-import)
3. [REST API 통합](#rest-api-통합)
4. [검색 최적화](#검색-최적화)
5. [테스트](#테스트)

## 기본 사용법

### 1. Repository + Service 초기화

```go
package main

import (
    postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
    "gorm.io/gorm"
)

func main() {
    var db *gorm.DB // DB 연결

    // Repository 생성
    repo := postalcodeapi.NewRepository(db)

    // Service 생성
    service := postalcodeapi.NewService(repo)

    // 사용
    results, _ := service.GetByZipCode("01000")
}
```

## 데이터 Import

### 1. 데이터 다운로드

먼저 우체국에서 최신 데이터를 다운로드합니다:

**다운로드 링크**: [우체국 우편번호 서비스](https://www.epost.go.kr/search/zipcode/areacdAddressDown.jsp)

1. 위 링크 접속
2. **"범위주소 DB"** 다운로드 후 압축해제
3. 파일 준비:
```bash
cp ~/Downloads/도로명주소*.txt data/road_address.txt
cp ~/Downloads/지번주소*.txt data/land_address.txt
```

### 2. Shell 스크립트로 Import (권장)

```bash
# 도로명주소 데이터 import
./scripts/import.sh \
    -file "data/road_address.txt" \
    -type road \
    -batch 1000

# 지번주소 데이터 import
./scripts/import.sh \
    -file "data/land_address.txt" \
    -type land \
    -batch 1000

# DSN 직접 지정
./scripts/import.sh \
    -dsn "user:pass@tcp(localhost:3306)/dbname" \
    -file "data/road_address.txt" \
    -type road
```

**장점**:
- ✅ 가장 간단한 사용법
- ✅ 별도 빌드 불필요
- ✅ 진행 상황 실시간 표시
- ✅ 에러 자동 처리 및 로깅

⚠️ **중요**: Import 시 기존 데이터가 자동으로 TRUNCATE되고 새 데이터로 대체됩니다.

### 3. 패키지 Importer 사용

프로그래밍 방식으로 import하려면:

```go
package main

import (
    "fmt"
    postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
    "gorm.io/gorm"
)

func main() {
    var db *gorm.DB // DB 연결

    repo := postalcodeapi.NewRepository(db)
    service := postalcodeapi.NewService(repo)
    importer := postalcodeapi.NewImporter(service)

    // 진행 상황 콜백
    progressFn := func(current, total int) {
        fmt.Printf("Progress: %d/%d (%.1f%%)\n",
            current, total, float64(current)/float64(total)*100)
    }

    // 도로명주소 import (기존 데이터 자동 TRUNCATE)
    result, err := importer.ImportFromFile(
        "data/road_address.txt",
        1000, // batch size
        progressFn,
    )

    // 지번주소 import (기존 데이터 자동 TRUNCATE)
    landResult, err := importer.ImportLandFromFile(
        "data/land_address.txt",
        1000,
        progressFn,
    )
}
```

💡 **Import 동작**: Import는 항상 기존 테이블 데이터를 TRUNCATE한 후 새 데이터를 삽입합니다.

### 4. CLI 도구 사용

```bash
# 빌드
go build -o postalcode-import cmd/postalcode-import/main.go

# 도로명주소 import
./postalcode-import \
    -file "data/road_address.txt" \
    -type road \
    -batch 1000

# 지번주소 import (DSN 직접 지정)
./postalcode-import \
    -dsn "user:pass@tcp(localhost:3306)/dbname" \
    -file "data/land_address.txt" \
    -type land \
    -batch 1000
```

**플래그 설명**:
- `-file`: 데이터 파일 경로 (필수)
- `-type`: 데이터 타입 - `road` 또는 `land` (필수)
- `-dsn`: MySQL DSN (선택, 없으면 .env 사용)
- `-batch`: 배치 크기 (기본: 1000)

### 파일 형식

파이프(`|`) 구분자 CSV 파일:
```
우편번호|시도|시도영문|시군구|시군구영문|읍면|읍면영문|도로명|도로명영문|지하여부|시작건물번호(주)|시작건물번호(부)|끝건물번호(주)|끝건물번호(부)|범위종류
01000|서울특별시|Seoul|강북구|Gangbuk-gu|||삼양로177길|Samyang-ro 177-gil|0|93|0|126|0|3
```

## REST API 통합

### 표준 http.ServeMux

```go
package main

import (
    "net/http"
    postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
    "gorm.io/gorm"
)

func main() {
    var db *gorm.DB

    // Setup
    repo := postalcodeapi.NewRepository(db)
    service := postalcodeapi.NewService(repo)
    handler := postalcodeapi.NewHandler(service)

    // 라우트 등록
    mux := http.NewServeMux()
    handler.RegisterRoutes(mux, "/api/v1/postal-codes")

    // 서버 시작
    http.ListenAndServe(":8080", mux)
}
```

### Gin 프레임워크 (권장)

#### 방법 1: GinHandler 사용 (추천)

```go
package main

import (
    "github.com/gin-gonic/gin"
    postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
    "gorm.io/gorm"
)

func main() {
    var db *gorm.DB

    repo := postalcodeapi.NewRepository(db)
    service := postalcodeapi.NewService(repo)

    // Gin 핸들러 생성 (Swagger 문서 포함!)
    handler := postalcodeapi.NewGinHandler(service)

    r := gin.Default()

    // 3줄로 모든 엔드포인트 등록 완료
    handler.RegisterGinRoutes(r.Group("/api/v1/postal-codes"))

    r.Run(":8080")
}
```

**GinHandler 장점**:
- ✅ Swagger 문서 자동 포함 (`--parseDependency`로 자동 통합)
- ✅ 3줄로 모든 엔드포인트 등록 완료
- ✅ 일관된 에러 처리 및 응답 형식
- ✅ 유지보수 용이 (패키지 업데이트 시 자동 반영)

**등록되는 엔드포인트**:
- `GET /api/v1/postal-codes/road/search` - 복합 조건 검색
- `GET /api/v1/postal-codes/road/zipcode/:code` - 우편번호로 조회
- `GET /api/v1/postal-codes/road/prefix/:prefix` - 우편번호 앞 3자리로 조회

#### 방법 2: 수동 핸들러 작성

수동으로 핸들러를 작성하려면 다음과 같이 할 수 있습니다:

```go
package main

import (
    "github.com/gin-gonic/gin"
    postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
    "gorm.io/gorm"
)

func main() {
    var db *gorm.DB

    repo := postalcodeapi.NewRepository(db)
    service := postalcodeapi.NewService(repo)

    r := gin.Default()

    // API 그룹
    api := r.Group("/api/v1/postal-codes")
    {
        api.GET("/zipcode/:code", func(c *gin.Context) {
            code := c.Param("code")
            results, err := service.GetByZipCode(code)
            if err != nil {
                c.JSON(400, gin.H{"error": err.Error()})
                return
            }
            c.JSON(200, gin.H{"data": results})
        })

        api.GET("/prefix/:prefix", func(c *gin.Context) {
            prefix := c.Param("prefix")
            results, _, err := service.GetByZipPrefix(prefix, 10, 0)
            if err != nil {
                c.JSON(400, gin.H{"error": err.Error()})
                return
            }
            c.JSON(200, gin.H{"data": results})
        })

        api.GET("/search", func(c *gin.Context) {
            var params postalcodeapi.SearchParams
            if err := c.ShouldBindQuery(&params); err != nil {
                c.JSON(400, gin.H{"error": err.Error()})
                return
            }
            results, total, err := service.Search(params)
            if err != nil {
                c.JSON(500, gin.H{"error": err.Error()})
                return
            }
            c.JSON(200, gin.H{
                "data":  results,
                "total": total,
            })
        })
    }

    r.Run(":8080")
}
```

⚠️ **주의**: 수동으로 작성하면 Swagger 문서가 자동으로 생성되지 않습니다. GinHandler 사용을 권장합니다.

### Echo 프레임워크

```go
package main

import (
    "github.com/labstack/echo/v4"
    postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
    "gorm.io/gorm"
)

func main() {
    var db *gorm.DB

    repo := postalcodeapi.NewRepository(db)
    service := postalcodeapi.NewService(repo)

    e := echo.New()

    // 라우트
    e.GET("/api/v1/postal/zipcode/:code", func(c echo.Context) error {
        code := c.Param("code")
        results, err := service.GetByZipCode(code)
        if err != nil {
            return c.JSON(400, map[string]string{"error": err.Error()})
        }
        return c.JSON(200, map[string]interface{}{"data": results})
    })

    e.Start(":8080")
}
```

## 검색 최적화

### 빠른 검색 (zip_prefix 사용)

```go
// ❌ 느림 - LIKE 연산
params := postalcodeapi.SearchParams{
    ZipCode: "010%", // LIKE 검색
}
results, _, _ := service.Search(params)

// ✅ 빠름 - 정확한 매칭 (3-5배 빠름)
results, _, _ := service.GetByZipPrefix("010", 10, 0)
```

### 페이징

```go
params := postalcodeapi.SearchParams{
    SidoName: "서울",
    Limit:    20,   // 페이지 크기
    Page:     3,    // 페이지 번호
}
results, total, _ := service.Search(params)

// 총 페이지 수 계산
totalPages := (total + int64(params.Limit) - 1) / int64(params.Limit)
```

### 복합 검색

```go
params := postalcodeapi.SearchParams{
    SidoName:    "서울특별시",
    SigunguName: "강북구",
    RoadName:    "삼양로",
    Limit:       10,
}
results, total, _ := service.Search(params)
```

## API 엔드포인트

### GET /postal-codes/zipcode/{zipCode}
우편번호로 정확히 조회

```bash
curl http://localhost:8080/api/v1/postal-codes/road/zipcode/01000
```

Response:
```json
{
  "success": true,
  "data": [...],
  "total": 5
}
```

### GET /postal-codes/prefix/{zipPrefix}
우편번호 앞 3자리로 빠른 검색

```bash
curl http://localhost:8080/api/v1/postal-codes/road/prefix/010
```

### GET /postal-codes/search
복합 검색

```bash
curl "http://localhost:8080/api/v1/postal-codes/road/search?sido_name=서울&limit=10&offset=0"
```

Query Parameters:
- `zip_code` - 우편번호 정확한 매칭
- `zip_prefix` - 우편번호 앞 3자리 정확한 매칭 (권장)
- `sido_name` - 시도명 부분 매칭
- `sigungu_name` - 시군구명 부분 매칭
- `road_name` - 도로명 부분 매칭
- `limit` - 결과 개수 (기본 100, 최대 1000)
- `offset` - 페이징 오프셋

## 성능 벤치마크

31만건 데이터 기준:

| 쿼리 | 실행시간 | 방법 |
|------|---------|------|
| `zip_prefix = '010'` | ~1-5ms | ✅ 권장 |
| `zip_code LIKE '010%'` | ~5-15ms | |
| `zip_code = '01000'` | ~1-3ms | ✅ 권장 |
| 복합 검색 (3개 조건) | ~10-30ms | |

## 테이블 자동 생성

### 방법 1: Migration CLI (권장)

가장 쉬운 방법은 제공되는 migration CLI 도구를 사용하는 것입니다:

```bash
# 빌드
go build -o postalcode-migrate cmd/postalcode-migrate/main.go

# .env 파일 사용 (권장)
./postalcode-migrate -cmd=up
./postalcode-migrate -cmd=status

# 또는 DSN 직접 지정
./postalcode-migrate -dsn="user:pass@tcp(localhost:3306)/dbname" -cmd=up

# 테이블 삭제
./postalcode-migrate -cmd=down

# 테이블 재생성 (삭제 후 생성)
./postalcode-migrate -cmd=fresh
```

**DSN 설정**:
- `-dsn` 플래그: 직접 지정 (우선순위 1)
- `.env` 파일: 자동 로드 (우선순위 2, configs/.env.example 참고)

**장점**:
- ✅ 간편한 사용법 (.env 파일 자동 로드)
- ✅ 테이블 상태 및 데이터 개수 확인
- ✅ 안전한 마이그레이션 관리
- ✅ 별도 코드 작성 불필요

### 방법 2: GORM AutoMigrate

프로그래밍 방식으로 테이블을 생성하려면:

```go
import (
    postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
    "gorm.io/gorm"
)

func main() {
    var db *gorm.DB

    // 테이블 자동 생성
    db.AutoMigrate(&postalcodeapi.PostalCodeRoad{}, &postalcodeapi.PostalCodeLand{})
}
```

### 방법 3: 수동 SQL

```bash
mysql -u user -p database < migrations/009_create_postal_code_roads_table.sql
mysql -u user -p database < migrations/010_create_postal_code_lands_table.sql
```

## 트러블슈팅

### Import 속도가 느림
- 배치 사이즈를 늘려보세요 (1000 → 5000)
- DB 인덱스가 생성되었는지 확인
- MySQL의 `innodb_flush_log_at_trx_commit` 설정 확인

### 메모리 부족
- 배치 사이즈를 줄이세요 (1000 → 500)
- 파일을 분할해서 여러 번 import

### 중복 데이터 에러
- Upsert 로직이 자동으로 처리합니다
- unique index 확인: `idx_postal_unique`

## 테스트

### 테스트 실행

```bash
# 전체 테스트 실행
go test ./...

# 특정 패키지 테스트
go test ./internal/repository
go test ./internal/service
go test ./internal/importer
go test ./internal/http

# 커버리지 포함
go test -cover ./...
```

### 테스트 구조

```
tests/
├── testdata/
│   ├── sample_road.txt    # 도로명주소 샘플 데이터
│   └── sample_land.txt    # 지번주소 샘플 데이터
├── integration_test.go    # 통합 테스트
internal/
├── repository/
│   └── repository_test.go # Repository 계층 테스트
├── service/
│   └── service_test.go    # Service 계층 테스트
├── importer/
│   └── importer_test.go   # Importer 테스트
└── http/
    ├── handler_test.go    # 표준 HTTP 핸들러 테스트
    └── gin_test.go        # Gin 핸들러 테스트
pkg/postalcode/
└── postalcode_test.go     # Public API 테스트
```

### 테스트 커버리지

- **Repository**: CRUD, 검색, 페이징, 에러 처리
- **Service**: 비즈니스 로직, 유효성 검사, Upsert 로직
- **Importer**: 파일 파싱, 배치 처리, 진행 상황 추적
- **HTTP Handler**: 엔드포인트, 요청/응답 검증, 에러 처리
- **Integration**: 전체 워크플로우, 복합 시나리오

### 테스트 데이터

테스트에 사용되는 샘플 데이터는 `tests/testdata/` 디렉토리에 있습니다:

- `sample_road.txt`: 도로명주소 샘플 (3개 레코드)
- `sample_land.txt`: 지번주소 샘플 (3개 레코드)

모든 테스트는 in-memory SQLite 데이터베이스를 사용하여 격리된 환경에서 실행됩니다.

## 라이센스

MIT
