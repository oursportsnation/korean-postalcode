# 다른 프로젝트에서 사용하기

이 문서는 `github.com/oursportsnation/korean-postalcode` 패키지를 다른 Go 프로젝트에서 사용하는 방법을 설명합니다.

## 방법 1: 기존 프로젝트에서 DB 연결 재사용

기존 프로젝트 내부에서는 **프로젝트의 DB 연결을 재사용**합니다.

### 핵심: postalcode는 `*gorm.DB`만 받음

postalcode 패키지는 설정이나 연결을 직접 관리하지 않습니다.
이미 연결된 `*gorm.DB`만 받기 때문에 **기존 프로젝트의 설정을 그대로 재사용** 가능합니다.

```go
package main

import (
    "your-project/internal/infrastructure/config"
    "your-project/internal/infrastructure/datastore"
    postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
)

func main() {
    // 1. 기존 프로젝트 설정 로드
    cfg, _ := config.LoadConfig("configs/config.yaml")

    // 2. 기존 프로젝트 DB 연결
    datastore.NewMySQLConnection(cfg.Database, cfg.Env)
    db := datastore.GetDB()

    // 3. postalcode 초기화 (기존 DB 재사용!)
    repo := postalcodeapi.NewRepository(db)  // 기존 프로젝트의 DB
    service := postalcodeapi.NewService(repo)

    // 4. 사용
    results, _, _ := service.GetByZipPrefix("010", 10, 0)
}
```

### Gin 프레임워크 통합

```go
import (
    "github.com/gin-gonic/gin"
    postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
)

func main() {
    // 기존 프로젝트 DB 연결 (위와 동일)
    repo := postalcodeapi.NewRepository(datastore.DB)
    service := postalcodeapi.NewService(repo)

    // Gin 핸들러 생성 (Swagger 문서 포함)
    handler := postalcodeapi.NewGinHandler(service)

    // Gin 라우터
    r := gin.Default()

    // 간단한 라우트 등록 - Swagger 문서 자동 포함!
    handler.RegisterGinRoutes(r.Group("/api/v1/postal-codes"))

    r.Run(":8080")
}
```

**GinHandler 장점**:
- ✅ Swagger 문서가 자동으로 포함됨
- ✅ 3줄로 모든 엔드포인트 등록 완료
- ✅ 일관된 에러 처리 및 응답 형식
- ✅ 유지보수 용이

### Swagger 문서 통합

postalcode 패키지의 swagger 주석을 외부 서비스의 swagger 문서에 자동으로 포함시킬 수 있습니다.

```bash
# 기존 프로젝트에서 swagger 문서 생성 시
swag init -g cmd/api/main.go -o docs/swagger --parseDependency --parseInternal
```

**핵심**: `--parseDependency` 플래그를 사용하면 import된 pkg/postalcode의 swagger 주석까지 자동으로 파싱됩니다.

**결과**:
- ✅ PostalCode API 엔드포인트가 기존 프로젝트 swagger 문서에 자동 포함
- ✅ PostalCodeRoad 모델이 스키마에 자동 포함
- ✅ SearchParams 구조체가 파라미터로 자동 포함
- ✅ 별도 문서 관리 불필요

**Swagger UI에서 확인**:
```
http://localhost:8080/swagger/index.html

Tags:
  - PostalCode (패키지에서 자동 포함됨)
    - GET /api/v1/postal-codes/{road|land}/search
    - GET /api/v1/postal-codes/{road|land}/zipcode/{code}
    - GET /api/v1/postal-codes/{road|land}/{prefix}
```

### 💡 핵심 포인트

✅ **설정 재사용**: 기존 프로젝트의 Database Config 그대로 사용
✅ **DB 재사용**: 기존 프로젝트의 `*gorm.DB` 그대로 전달
✅ **AutoMigrate 공유**: 기존 프로젝트의 AutoMigrate에 추가
✅ **트랜잭션 공유**: 같은 DB 연결 풀 사용
✅ **Swagger 통합**: `--parseDependency` 플래그로 문서 자동 포함

**postalcode.Config는 standalone 사용을 위한 optional helper일 뿐입니다!**

## 방법 2: 다른 Go 프로젝트에서 (GitHub)

### 1단계: 패키지 설치

```bash
cd my-service
go get github.com/oursportsnation/korean-postalcode
```

### 2단계: 코드에서 사용

**my-service/go.mod**:
```go
module my-service

go 1.21

require (
    github.com/oursportsnation/korean-postalcode v1.0.0
    gorm.io/gorm v1.25.0
)
```

**my-service/main.go**:
```go
package main

import (
    postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
    "gorm.io/gorm"
)

func main() {
    var db *gorm.DB

    repo := postalcodeapi.NewRepository(db)
    service := postalcodeapi.NewService(repo)

    results, _ := service.GetByZipCode("01000")
}
```

## 방법 3: 같은 모노레포 내 다른 서비스 (replace 사용)

```
my-monorepo/
  shared/               # 공유 패키지
    pkg/postalcode/
  service-a/            # 서비스 A
    main.go
  service-b/            # 서비스 B
    main.go
```

**service-a/go.mod**:
```go
module service-a

go 1.21

require (
    github.com/oursportsnation/korean-postalcode v1.0.0
    gorm.io/gorm v1.25.0
)

replace github.com/oursportsnation/korean-postalcode => ../shared
```

**my-service/main.go**:
```go
package main

import (
    postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
    "gorm.io/gorm"
)

func main() {
    var db *gorm.DB

    repo := postalcodeapi.NewRepository(db)
    service := postalcodeapi.NewService(repo)

    results, _, _ := service.GetByZipPrefix("010", 10, 0)
}
```

## 실제 사용 예제

### 예제 1: 주소 검증 마이크로서비스

```go
package main

import (
    "github.com/gin-gonic/gin"
    postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
    "gorm.io/gorm"
)

type AddressValidator struct {
    postalService postalcodeapi.Service
}

func NewAddressValidator(db *gorm.DB) *AddressValidator {
    repo := postalcodeapi.NewRepository(db)
    service := postalcodeapi.NewService(repo)

    return &AddressValidator{
        postalService: service,
    }
}

func (v *AddressValidator) ValidateAddress(zipCode, roadName string, buildingNo int) (bool, error) {
    // 우편번호로 조회
    roads, err := v.postalService.GetByZipCode(zipCode)
    if err != nil {
        return false, err
    }

    // 도로명과 건물번호 검증
    for _, road := range roads {
        if road.RoadName == roadName {
            // 건물번호 범위 체크
            if buildingNo >= road.StartBuildingMain {
                if road.EndBuildingMain == nil || buildingNo <= *road.EndBuildingMain {
                    return true, nil
                }
            }
        }
    }

    return false, nil
}

func main() {
    var db *gorm.DB // DB 연결

    validator := NewAddressValidator(db)

    r := gin.Default()
    r.POST("/validate", func(c *gin.Context) {
        var req struct {
            ZipCode    string `json:"zip_code"`
            RoadName   string `json:"road_name"`
            BuildingNo int    `json:"building_no"`
        }

        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(400, gin.H{"error": err.Error()})
            return
        }

        valid, err := validator.ValidateAddress(req.ZipCode, req.RoadName, req.BuildingNo)
        if err != nil {
            c.JSON(500, gin.H{"error": err.Error()})
            return
        }

        c.JSON(200, gin.H{"valid": valid})
    })

    r.Run(":8080")
}
```

### 예제 2: 배송지 관리 서비스

```go
package main

import (
    postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
    "gorm.io/gorm"
)

type DeliveryService struct {
    postalService postalcodeapi.Service
}

func NewDeliveryService(db *gorm.DB) *DeliveryService {
    repo := postalcodeapi.NewRepository(db)
    service := postalcodeapi.NewService(repo)

    return &DeliveryService{
        postalService: service,
    }
}

// GetDeliveryRegion은 우편번호로 배송 지역을 찾습니다.
func (s *DeliveryService) GetDeliveryRegion(zipCode string) (string, error) {
    roads, err := s.postalService.GetByZipCode(zipCode)
    if err != nil {
        return "", err
    }

    if len(roads) == 0 {
        return "unknown", nil
    }

    // 시도명으로 배송 지역 결정
    sido := roads[0].SidoName

    switch sido {
    case "서울특별시", "경기도", "인천광역시":
        return "수도권", nil
    case "부산광역시", "대구광역시", "울산광역시", "경상남도", "경상북도":
        return "영남권", nil
    default:
        return "기타", nil
    }
}

// AutocompleteAddress는 입력된 주소로 자동완성 목록을 제공합니다.
func (s *DeliveryService) AutocompleteAddress(query string) ([]postalcodeapi.PostalCodeRoad, error) {
    params := postalcodeapi.SearchParams{
        RoadName: query,
        Limit:    10,
    }

    results, _, err := s.postalService.Search(params)
    return results, err
}
```

### 예제 3: 주소 검색 API

```go
package main

import (
    "net/http"
    postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
    "gorm.io/gorm"
)

func main() {
    var db *gorm.DB // DB 연결

    // PostalCode 서비스 설정
    repo := postalcodeapi.NewRepository(db)
    service := postalcodeapi.NewService(repo)
    handler := postalcodeapi.NewHandler(service)

    // API 라우트 등록
    mux := http.NewServeMux()

    // PostalCode API 마운트
    handler.RegisterRoutes(mux, "/api/v1/postal")

    // 추가 커스텀 엔드포인트
    mux.HandleFunc("/api/v1/regions", func(w http.ResponseWriter, r *http.Request) {
        prefix := r.URL.Query().Get("prefix")
        roads, _, _ := service.GetByZipPrefix(prefix, 100, 0)

        // 시도별로 그룹화
        regions := make(map[string]int)
        for _, road := range roads {
            regions[road.SidoName]++
        }

        // JSON 응답
        // ...
    })

    http.ListenAndServe(":8080", mux)
}
```

## 테이블 자동 생성

다른 프로젝트에서 PostalCode 패키지를 사용하기 전에 테이블을 생성해야 합니다.

### 방법 1: Migration CLI 사용 (권장)

패키지가 제공하는 Migration CLI를 사용하면 가장 쉽게 테이블을 관리할 수 있습니다:

```bash
# PostalCode 저장소에서 빌드
cd /path/to/korean-postalcode
go build -o postalcode-migrate cmd/postalcode-migrate/main.go

# 다른 프로젝트의 DB에 테이블 생성
./postalcode-migrate \
    -dsn="user:pass@tcp(localhost:3306)/your_project_db" \
    -cmd=up

# 테이블 상태 확인
./postalcode-migrate \
    -dsn="user:pass@tcp(localhost:3306)/your_project_db" \
    -cmd=status
```

**사용 가능한 명령어**:
- `up`: 테이블 생성 (postal_code_roads, postal_code_lands)
- `down`: 테이블 삭제
- `fresh`: 테이블 재생성 (삭제 후 생성)
- `status`: 테이블 상태 및 데이터 개수 확인

**장점**:
- ✅ 별도 코드 작성 불필요
- ✅ 테이블 상태 실시간 확인
- ✅ 안전한 마이그레이션 관리
- ✅ 여러 환경(dev/staging/prod)에 동일하게 적용 가능

### 방법 2: AutoMigrate 사용

프로그래밍 방식으로 테이블을 생성하려면:

```go
package main

import (
    postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
    "gorm.io/gorm"
)

func main() {
    var db *gorm.DB // DB 연결

    // 테이블 자동 생성 (도로명주소 + 지번주소)
    db.AutoMigrate(&postalcodeapi.PostalCodeRoad{}, &postalcodeapi.PostalCodeLand{})

    // 이후 사용
    repo := postalcodeapi.NewRepository(db)
    // ...
}
```

**권장 사용 시기**:
- 애플리케이션 시작 시 자동으로 테이블을 생성하고 싶을 때
- 기존 프로젝트의 migration 시스템에 통합하고 싶을 때

## 버전 관리

### Semantic Versioning 사용

```bash
# korean-postalcode 저장소에서
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

### 다른 프로젝트에서 특정 버전 사용

```bash
# 특정 버전
go get github.com/oursportsnation/korean-postalcode@v1.0.0

# 최신 버전
go get github.com/oursportsnation/korean-postalcode@latest

# 특정 커밋
go get github.com/oursportsnation/korean-postalcode@commit-hash
```

## 트러블슈팅

### Import Path 문제

```go
// ❌ 잘못된 import
import "pkg/postalcode"

// ✅ 올바른 import
import postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
```

### Database Connection 문제

패키지는 DB 연결을 직접 관리하지 않습니다. 사용하는 쪽에서 GORM DB 인스턴스를 제공해야 합니다.

```go
// DB 연결은 사용하는 쪽에서 관리
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

// 패키지에 전달
repo := postalcodeapi.NewRepository(db)
```

### 패키지 설치 오류

**문제**: `go get` 시 실패

**해결**:
```bash
# Go 모듈 캐시 정리
go clean -modcache

# 다시 설치
go get github.com/oursportsnation/korean-postalcode
```

## 📚 관련 문서

- [패키지 README](../README.md) - 패키지 개요 및 설치
- [REST API 엔드포인트 가이드](./API.md) - 완전한 API 문서
- [완전한 사용 가이드](./USAGE.md) - Repository/Service 사용법
- [실행 가능한 예제](../examples/README.md) - 코드 예제

## 라이센스

MIT
