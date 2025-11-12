# Korean PostalCode

한국 우편번호, 도로명주소 및 지번주소 데이터를 관리하는 재사용 가능한 Go 패키지입니다.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/oursportsnation/korean-postalcode.svg)](https://pkg.go.dev/github.com/oursportsnation/korean-postalcode)

## ✨ 특징

- ✅ **도로명주소 & 지번주소 지원**: 두 가지 주소 체계 모두 지원
- ✅ **재사용 가능**: 다른 Go 서비스에서 import하여 사용
- ✅ **REST API 제공**: Gin 기반 HTTP API 서버 내장
- ✅ **레이어 분리**: Repository, Service, Handler 패턴
- ✅ **고성능 검색**: 우편번호 3자리 prefix 인덱스 최적화 (3-5배 빠름)
- ✅ **배치 처리**: 대량 데이터 import 지원 (도로명/지번 모두)
- ✅ **표준화된 에러**: 커스텀 에러 타입 제공
- ✅ **완벽한 테스트**: 100+ 테스트로 검증된 안정성

## 📦 설치

```bash
go get github.com/oursportsnation/korean-postalcode
```

## 🚀 빠른 시작

### 1. 테이블 생성 (Migration)

```bash
# CLI 도구 빌드
go build -o postalcode-migrate cmd/postalcode-migrate/main.go

# 방법 1: .env 파일 사용 (권장)
./postalcode-migrate -cmd=up

# 방법 2: DSN 직접 지정
./postalcode-migrate -dsn="user:pass@tcp(localhost:3306)/dbname" -cmd=up

# 상태 확인
./postalcode-migrate -cmd=status
```

**Migration 명령어**:
- `up`: 테이블 생성
- `down`: 테이블 삭제
- `fresh`: 테이블 재생성 (삭제 후 생성)
- `status`: 테이블 상태 및 데이터 개수 확인

**DSN 설정**:
- `-dsn` 플래그 사용 (우선순위 1)
- `.env` 파일 자동 로드 (우선순위 2)

### 2. 데이터 Import

```bash
# Shell 스크립트 사용 (권장)
./scripts/import.sh \
    "user:pass@tcp(localhost:3306)/dbname" \
    "data/postal_codes.txt" \
    1000
```

### 3. 기본 사용

```go
import (
    "github.com/oursportsnation/korean-postalcode"
    postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

// DB 연결
dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})

// Repository & Service 생성
repo := postalcodeapi.NewRepository(db)
service := postalcodeapi.NewService(repo)

// 도로명주소 조회
roadResults, _ := service.GetByZipCode("01000")

// 지번주소 조회
landResults, _ := service.GetLandByZipCode("25627")
```

### 4. 환경 설정 사용 (Standalone)

```go
import (
    "github.com/oursportsnation/korean-postalcode"
    postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
)

// .env 파일에서 설정 로드
cfg, _ := postalcode.LoadConfig()

// DB 연결
db, _ := gorm.Open(mysql.Open(cfg.Database.GetDSN()), &gorm.Config{})

// Repository & Service 생성
repo := postalcodeapi.NewRepository(db)
service := postalcodeapi.NewService(repo)
```

### 5. 기존 프로젝트 DB 재사용

```go
import (
    postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
    "your-project/internal/database" // 기존 프로젝트의 DB
)

// 기존 프로젝트의 DB를 그대로 재사용!
repo := postalcodeapi.NewRepository(database.DB)
service := postalcodeapi.NewService(repo)

// 사용
results, _ := service.GetByZipCode("01000")
```

💡 **핵심**: `*gorm.DB`만 받으므로 어떤 프로젝트의 DB든 재사용 가능!

## 🌐 REST API 서버

### Gin API 서버 실행 (권장)

```bash
# 빌드
cd cmd/postalcode-api
go build -o postalcode-api

# .env 파일 설정 후 실행
./postalcode-api

# 또는 플래그로 직접 설정
./postalcode-api -dsn "user:pass@tcp(localhost:3306)/dbname" -port 8080
```

**자동으로 제공되는 기능**:
- ✅ 도로명주소 & 지번주소 API 엔드포인트
- ✅ Swagger UI 문서 (http://localhost:8080/swagger/index.html)
- ✅ CORS 지원
- ✅ Graceful shutdown
- ✅ Health check
- ✅ 자동 DB 마이그레이션

자세한 내용은 [cmd/postalcode-api/README.md](cmd/postalcode-api/README.md) 참조

### Swagger API 문서

API 서버 실행 후 브라우저에서 접속:
```
http://localhost:8080/swagger/index.html
```

**Swagger 문서 재생성**:
```bash
# swag CLI 설치 (최초 1회)
go install github.com/swaggo/swag/cmd/swag@v1.8.12

# 문서 재생성
swag init -g cmd/postalcode-api/main.go -o docs/swagger --parseDependency --parseInternal
```

### 프로그래밍 방식으로 사용

#### Gin 프레임워크

```go
import (
    "github.com/gin-gonic/gin"
    postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
)

service := postalcodeapi.NewService(repo)
router := gin.Default()
postalcodeapi.RegisterGinRoutes(service, router.Group("/api/v1/postal-codes"))
router.Run(":8080")
```

#### 표준 HTTP

```go
import (
    "net/http"
    postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
)

service := postalcodeapi.NewService(repo)
mux := http.NewServeMux()
postalcodeapi.RegisterHTTPRoutes(service, mux, "/api/v1/postal-codes")
http.ListenAndServe(":8080", mux)
```

## 🔍 API 엔드포인트

### 도로명주소 API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/road/zipcode/{code}` | GET | 우편번호로 정확히 조회 (5자리) |
| `/road/prefix/{prefix}` | GET | 우편번호 앞 3자리로 빠른 검색 (권장) |
| `/road/search` | GET | 복합 검색 (시도, 시군구, 도로명) |

**Example:**
```bash
curl http://localhost:8080/api/v1/postal-codes/road/zipcode/01000
curl http://localhost:8080/api/v1/postal-codes/road/prefix/010
curl "http://localhost:8080/api/v1/postal-codes/road/search?sido_name=서울&limit=10"
```

### 지번주소 API

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/land/zipcode/{code}` | GET | 우편번호로 지번주소 조회 (5자리) |
| `/land/prefix/{prefix}` | GET | 우편번호 앞 3자리로 빠른 검색 (권장) |
| `/land/search` | GET | 복합 검색 (시도, 시군구, 읍면동, 리명) |

**Example:**
```bash
curl http://localhost:8080/api/v1/postal-codes/land/zipcode/25627
curl http://localhost:8080/api/v1/postal-codes/land/prefix/256
curl "http://localhost:8080/api/v1/postal-codes/land/search?sido_name=강원&eupmyeondong_name=강동면"
```

## 📊 데이터 Import

### 1. 데이터 다운로드 (우체국)

먼저 우체국 사이트에서 최신 우편번호 데이터를 다운로드합니다:

**다운로드 링크**: [우체국 우편번호 서비스](https://www.epost.go.kr/search/zipcode/areacdAddressDown.jsp)

**다운로드 방법**:
1. 위 링크 접속
2. **"범위주소 DB"** 다운로드 후 압축해제

**파일 준비**:
```bash
# 프로젝트의 data 디렉토리에 다운로드한 파일 복사
cp ~/Downloads/도로명주소*.txt data/road_address.txt
cp ~/Downloads/지번주소*.txt data/land_address.txt
```

💡 **참고**:
- 우체국 사이트의 파일명은 날짜별로 다를 수 있습니다 (예: `20251111_도로명주소.txt`)
- 파일 형식은 파이프(`|`) 구분자를 사용하는 TXT 파일입니다
- 파일 크기가 클 수 있으므로 (수백 MB), 다운로드에 시간이 걸릴 수 있습니다

### 2. Shell 스크립트로 Import (권장)

가장 쉬운 방법은 제공되는 shell 스크립트를 사용하는 것입니다:

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

# DSN 직접 지정하는 경우
./scripts/import.sh \
    -dsn "user:pass@tcp(localhost:3306)/dbname" \
    -file "data/road_address.txt" \
    -type road \
    -batch 1000
```

**스크립트 자동 기능**:
- ✅ 파일 존재 확인
- ✅ 파일 정보 출력 (크기, 라인 수)
- ✅ 바이너리 자동 빌드
- ✅ 진행 상황 표시
- ✅ 성공/실패 결과 출력

⚠️ **중요**: Import 시 **기존 데이터가 자동으로 삭제**(TRUNCATE)되고 새 데이터로 대체됩니다. 이는 우체국 데이터가 전체 스냅샷 방식으로 제공되기 때문입니다.

### 3. CLI 도구로 Import

수동으로 빌드하여 사용:

```bash
cd cmd/postalcode-import
go build -o postalcode-import

# 도로명주소 데이터 import (.env 파일 사용)
./postalcode-import \
    -file "data/road_address.txt" \
    -type road \
    -batch 1000

# 지번주소 데이터 import (DSN 직접 지정)
./postalcode-import \
    -dsn "user:pass@tcp(localhost:3306)/dbname" \
    -file "data/land_address.txt" \
    -type land \
    -batch 1000
```

**플래그 설명**:
- `-file`: 데이터 파일 경로 (필수)
- `-type`: 데이터 타입 - `road` (도로명주소) 또는 `land` (지번주소) (필수)
- `-dsn`: MySQL DSN (선택, 없으면 .env 파일 사용)
- `-batch`: 배치 처리 크기 (기본값: 1000)

⚠️ **주의**: Import는 항상 기존 데이터를 TRUNCATE한 후 새 데이터를 삽입합니다.

### 4. 프로그래밍 방식으로 Import

```go
import postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"

importer := postalcodeapi.NewImporter(service)

progressFn := func(current, total int) {
    fmt.Printf("Progress: %d/%d\n", current, total)
}

// 도로명주소 import (기존 데이터 자동 TRUNCATE)
result, err := importer.ImportFromFile("road_data.txt", 1000, progressFn)

// 지번주소 import (기존 데이터 자동 TRUNCATE)
landResult, err := importer.ImportLandFromFile("land_data.txt", 1000, progressFn)
```

💡 **Import 동작**:
- Import는 항상 기존 테이블 데이터를 TRUNCATE한 후 새 데이터를 삽입합니다
- 도로명주소(`ImportFromFile`)와 지번주소(`ImportLandFromFile`)는 각각 독립적인 테이블을 사용합니다
- 부분 업데이트가 필요한 경우 `service.Upsert()` 또는 `service.BatchUpsert()` 메서드를 사용하세요

## 🗄️ 데이터베이스 설정

### AutoMigrate (권장)

```go
import "github.com/oursportsnation/korean-postalcode"

// 도로명주소 및 지번주소 테이블 자동 생성
db.AutoMigrate(&postalcode.PostalCodeRoad{}, &postalcode.PostalCodeLand{})
```

### 수동 SQL

```bash
# 도로명주소 테이블
mysql -u user -p database < migrations/create_postal_code_roads.sql

# 지번주소 테이블
mysql -u user -p database < migrations/create_postal_code_lands.sql
```

## ⚡ 성능

31만건 데이터 기준:

| 검색 방법 | 실행시간 | 인덱스 |
|-----------|---------|--------|
| `zip_prefix = '010'` | ~1-5ms | idx_zip_prefix ✅ |
| `zip_code LIKE '010%'` | ~5-15ms | idx_zipcode |
| `zip_code = '01000'` | ~1-3ms | idx_zipcode ✅ |

**권장**: 우편번호 앞 3자리 검색은 `GetByZipPrefix()` 사용

## 🛠️ 에러 처리

```go
import (
    "errors"
    "github.com/oursportsnation/korean-postalcode"
)

results, err := service.GetByZipCode("01000")
if err != nil {
    switch {
    case errors.Is(err, postalcode.ErrNotFound):
        // Handle not found
    case errors.Is(err, postalcode.ErrInvalidZipCode):
        // Handle invalid format
    default:
        // Handle other errors
    }
}
```

## 📚 문서

- **[Swagger UI](http://localhost:8080/swagger/index.html)** - 인터랙티브 API 문서 (서버 실행 필요)
- **[API.md](docs/API.md)** - REST API 엔드포인트 완전 가이드
- **[USAGE.md](docs/USAGE.md)** - Repository/Service 사용 가이드
- **[INTEGRATION.md](docs/INTEGRATION.md)** - 다른 프로젝트 통합 방법
- **[examples/](examples/)** - 실행 가능한 코드 예제

## 📦 패키지 구조

```
korean-postalcode/
├── config.go              # 설정 관리 (공개 API)
├── errors.go              # 표준화된 에러 (공개 API)
├── models.go              # 데이터 모델 (공개 API)
├── internal/              # 비공개 구현
│   ├── repository/        # DB 접근 레이어
│   │   └── repository.go  # Repository 구현
│   ├── service/           # 비즈니스 로직
│   │   └── service.go     # Service 구현
│   ├── importer/          # 파일 Import 기능
│   │   └── importer.go    # Importer 구현
│   └── http/              # HTTP API 핸들러
│       ├── handler.go     # 표준 HTTP 핸들러
│       └── gin.go         # Gin 핸들러
├── pkg/                   # 공개 래퍼 API
│   └── postalcode/        # 편의 팩토리 함수
│       └── postalcode.go  # NewRepository, NewService, etc.
├── cmd/                   # CLI 도구
│   ├── postalcode-api/    # Gin API 서버
│   └── postalcode-import/ # 데이터 import 도구
├── docs/                  # 문서
│   ├── API.md             # API 가이드
│   ├── USAGE.md           # 사용 가이드
│   ├── INTEGRATION.md     # 통합 가이드
│   └── swagger/           # Swagger 문서
│       ├── docs.go        # Swagger 생성 파일
│       ├── swagger.json   # Swagger JSON
│       └── swagger.yaml   # Swagger YAML
├── migrations/            # SQL 마이그레이션
│   ├── create_postal_code_roads.sql  # 도로명주소 테이블
│   └── create_postal_code_lands.sql  # 지번주소 테이블
├── data/                  # 데이터 파일
│   ├── 20251111_road_name.txt       # 도로명주소 데이터
│   └── 20251111_land_rot.txt        # 지번주소 데이터
├── examples/              # 코드 예제
│   ├── basic/             # 기본 사용법
│   ├── api/               # REST API 서버
│   └── import/            # 데이터 import
├── scripts/               # Shell 스크립트
└── .env.example           # 환경변수 예제
```

### 패키지 설계 원칙

- **internal/**: 모든 구현 세부사항. 외부에서 직접 import 불가 (Go의 internal 패키지 규칙)
  - `repository/`: DB 접근 레이어
  - `service/`: 비즈니스 로직
  - `importer/`: 파일 import 기능
  - `http/`: HTTP 핸들러 (Gin & 표준 HTTP)
- **pkg/postalcode/**: 공개 API 진입점. 팩토리 함수와 라우트 등록 함수 제공
- **루트 패키지**: 공개 데이터 모델, 설정, 에러 타입만 노출

**완전한 캡슐화**: 사용자는 `pkg/postalcode`의 함수만 사용하며, 구현 세부사항은 완전히 숨겨집니다.

## 🧪 테스트

프로젝트는 100개 이상의 테스트로 완벽하게 검증되었습니다:

```bash
# 전체 테스트 실행
go test ./...

# 상세 출력과 함께 실행
go test -v ./...

# 커버리지 확인
go test -cover ./...
```

### 테스트 구조

```
tests/
├── internal/
│   ├── repository/
│   │   └── repository_test.go    # DB 접근 레이어 테스트
│   ├── service/
│   │   └── service_test.go       # 비즈니스 로직 테스트
│   ├── importer/
│   │   └── importer_test.go      # 파일 import 테스트
│   └── http/
│       ├── handler_test.go       # 표준 HTTP 핸들러 테스트
│       └── gin_test.go           # Gin 핸들러 테스트
├── pkg/postalcode/
│   └── postalcode_test.go        # 공개 API 테스트
├── tests/
│   ├── integration_test.go       # 통합 테스트
│   └── testdata/                 # 테스트 데이터
│       ├── sample_road.txt       # 도로명주소 샘플
│       └── sample_land.txt       # 지번주소 샘플
└── examples/
    └── example_test.go           # 사용 예제 테스트
```

### 테스트 커버리지

- ✅ **Repository Layer**: CRUD, 검색, 페이지네이션, 배치 작업
- ✅ **Service Layer**: 유효성 검증, 비즈니스 로직, 자동 zip prefix 추출
- ✅ **HTTP Handlers**: 모든 엔드포인트, HTTP 메서드, 쿼리 파라미터
- ✅ **Importer**: 파일 파싱, 에러 처리, 진행 상황 추적
- ✅ **Integration**: 전체 워크플로우, 에러 전파, 복잡한 검색
- ✅ **Public API**: 팩토리 함수, 라우트 등록, 인터페이스 호환성

모든 테스트는 in-memory SQLite를 사용하여 빠르고 격리된 환경에서 실행됩니다.

## 🤝 기여

Issues와 Pull Requests를 환영합니다!

## 📝 라이센스

MIT License - 자세한 내용은 [LICENSE](LICENSE) 파일을 참조하세요.

---

**Made with ❤️ for Korean Address Management**
