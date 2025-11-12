# PostalCode API 엔드포인트 가이드

한국 우편번호 및 도로명주소 검색 REST API 완전 가이드입니다.

## 📋 개요

이 API는 행정안전부 도로명주소 데이터(31만건+)를 기반으로 우편번호 및 도로명주소 검색 기능을 제공합니다.

### 주요 특징
- ✅ 31만건+ 행정안전부 도로명주소 데이터
- ✅ 우편번호 앞 3자리 prefix 인덱스 최적화 (3-5배 빠름)
- ✅ Repository/Service/Handler 레이어 분리
- ✅ Swagger 문서 자동 통합 지원
- ✅ 표준 HTTP REST API (http.ServeMux 또는 Gin 호환)

### 기본 URL 구조

패키지는 표준 `http.ServeMux` 또는 Gin 라우터에 마운트 가능합니다:

```go
import (
    postalcodeapi "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
)

// 표준 http.ServeMux
handler := postalcodeapi.NewHandler(service)
mux := http.NewServeMux()
handler.RegisterRoutes(mux, "/api/v1/postal-codes")

// Gin 프레임워크 (권장)
handler := postalcodeapi.NewGinHandler(service)
r := gin.Default()
handler.RegisterGinRoutes(r.Group("/api/v1/postal-codes"))
```

**기본 경로**: `/api/v1/postal-codes` (권장)

## 🔍 API 엔드포인트

### 1. 정확한 우편번호 조회

**엔드포인트**: `GET /api/v1/postal-codes/road/zipcode/{code}`

**목적**: 5자리 우편번호로 정확히 매칭되는 도로명주소 조회

**경로 파라미터**:
| 파라미터 | 타입 | 필수 | 설명 |
|---------|------|-----|------|
| `code` | string | Yes | 5자리 우편번호 |

**요청 예시**:
```bash
curl http://localhost:8080/api/v1/postal-codes/road/zipcode/01000
```

**응답 예시** (200 OK):
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "zip_code": "01000",
      "zip_prefix": "010",
      "sido_name": "서울특별시",
      "sido_name_en": "Seoul",
      "sigungu_name": "강북구",
      "sigungu_name_en": "Gangbuk-gu",
      "eupmyeon_name": "",
      "eupmyeon_name_en": "",
      "road_name": "삼양로177길",
      "road_name_en": "Samyang-ro 177-gil",
      "is_underground": false,
      "start_building_main": 93,
      "start_building_sub": 0,
      "end_building_main": 126,
      "end_building_sub": 0,
      "range_type": 3,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 1
}
```

**에러 응답**:
```json
// 400 Bad Request - 잘못된 형식
{
  "success": false,
  "error": "invalid zip code format"
}

// 404 Not Found - 우편번호를 찾을 수 없음
{
  "success": false,
  "error": "postal code not found"
}
```

---

### 2. 우편번호 prefix 빠른 검색 (권장)

**엔드포인트**: `GET /api/v1/postal-codes/road/prefix/{prefix}`

**목적**: 우편번호 앞 3자리로 빠른 검색

**특징**:
- 인덱스 최적화로 3-5배 빠른 검색 성능
- 대량 데이터 조회 시 권장

**경로 파라미터**:
| 파라미터 | 타입 | 필수 | 설명 |
|---------|------|-----|------|
| `prefix` | string | Yes | 우편번호 앞 3자리 |

**요청 예시**:
```bash
curl http://localhost:8080/api/v1/postal-codes/road/prefix/010
```

**응답 예시** (200 OK):
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "zip_code": "01000",
      "zip_prefix": "010",
      "sido_name": "서울특별시",
      "sigungu_name": "강북구",
      "road_name": "삼양로177길",
      "start_building_main": 93,
      "end_building_main": 126
    },
    {
      "id": 2,
      "zip_code": "01001",
      "zip_prefix": "010",
      "sido_name": "서울특별시",
      "sigungu_name": "강북구",
      "road_name": "삼양로173길",
      "start_building_main": 1,
      "end_building_main": 50
    }
    // ... 더 많은 결과
  ],
  "total": 1234
}
```

**에러 응답**:
```json
// 400 Bad Request - 잘못된 형식
{
  "success": false,
  "error": "invalid zip prefix format: must be 3 digits"
}
```

---

### 3. 복합 검색

**엔드포인트**: `GET /api/v1/postal-codes/road/search`

**목적**: 시도, 시군구, 도로명 등 여러 조건으로 유연한 검색

**쿼리 파라미터**:
| 파라미터 | 타입 | 필수 | 설명 | 예시 |
|---------|------|-----|------|------|
| `zip_code` | string | No | 우편번호 (5자리 정확 매칭) | `01000` |
| `zip_prefix` | string | No | 우편번호 앞 3자리 (권장, 빠름) | `010` |
| `sido_name` | string | No | 시도명 (부분 매칭) | `서울특별시` 또는 `서울` |
| `sigungu_name` | string | No | 시군구명 (부분 매칭) | `강북구` 또는 `강북` |
| `road_name` | string | No | 도로명 (부분 매칭) | `삼양로` |
| `limit` | int | No | 결과 개수 제한 (기본 100, 최대 1000) | `100` |
| `offset` | int | No | 페이징 오프셋 (기본 0) | `0` |

**사용 시나리오**:

#### 1) 시도명으로 검색
```bash
curl "http://localhost:8080/api/v1/postal-codes/road/search?sido_name=서울&limit=10"
```

#### 2) 복합 조건 검색
```bash
curl "http://localhost:8080/api/v1/postal-codes/road/search?sido_name=서울&sigungu_name=강북구&road_name=삼양로"
```

#### 3) prefix로 빠른 검색 후 필터링
```bash
# 추천: prefix를 먼저 사용하여 범위를 좁힌 후 추가 필터링
curl "http://localhost:8080/api/v1/postal-codes/road/search?zip_prefix=010&sigungu_name=강북구"
```

#### 4) 페이징
```bash
# 첫 페이지
curl "http://localhost:8080/api/v1/postal-codes/road/search?sido_name=서울&limit=50&offset=0"

# 두 번째 페이지
curl "http://localhost:8080/api/v1/postal-codes/road/search?sido_name=서울&limit=50&offset=50"
```

**응답 예시** (200 OK):
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "zip_code": "01000",
      "zip_prefix": "010",
      "sido_name": "서울특별시",
      "sigungu_name": "강북구",
      "road_name": "삼양로177길",
      "start_building_main": 93,
      "end_building_main": 126,
      "is_underground": false
    },
    {
      "id": 2,
      "zip_code": "01001",
      "zip_prefix": "010",
      "sido_name": "서울특별시",
      "sigungu_name": "강북구",
      "road_name": "삼양로173길",
      "start_building_main": 1,
      "end_building_main": 50,
      "is_underground": false
    }
    // ... 더 많은 결과
  ],
  "total": 25
}
```

**에러 응답**:
```json
// 400 Bad Request - 잘못된 파라미터
{
  "success": false,
  "error": "invalid search parameters"
}

// 500 Internal Server Error
{
  "success": false,
  "error": "internal server error"
}
```

---

## 🏠 지번주소 API

지번주소 조회를 위한 REST API 엔드포인트입니다.

### 1. 정확한 우편번호로 지번주소 조회

**엔드포인트**: `GET /api/v1/postal-codes/land/zipcode/{code}`

**목적**: 5자리 우편번호로 지번주소 조회

**경로 파라미터**:
| 파라미터 | 타입 | 필수 | 설명 |
|---------|------|-----|------|
| `code` | string | Yes | 5자리 우편번호 |

**요청 예시**:
```bash
curl http://localhost:8080/api/v1/postal-codes/land/zipcode/25627
```

**응답 예시** (200 OK):
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "zip_code": "25627",
      "zip_prefix": "256",
      "sido_name": "강원특별자치도",
      "sido_name_en": "Gangwon-do",
      "sigungu_name": "강릉시",
      "sigungu_name_en": "Gangneung-si",
      "eupmyeondong_name": "강동면",
      "eupmyeondong_name_en": "Gangdong-myeon",
      "ri_name": "모전리",
      "haengjeongdong_name": "강동면",
      "is_mountain": false,
      "start_jibun_main": 21,
      "start_jibun_sub": 0,
      "end_jibun_main": 198,
      "end_jibun_sub": 0,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "total": 2
}
```

---

### 2. 우편번호 prefix로 지번주소 빠른 검색

**엔드포인트**: `GET /api/v1/postal-codes/land/prefix/{prefix}`

**목적**: 우편번호 앞 3자리로 지번주소 빠른 검색

**경로 파라미터**:
| 파라미터 | 타입 | 필수 | 설명 |
|---------|------|-----|------|
| `prefix` | string | Yes | 우편번호 앞 3자리 |

**요청 예시**:
```bash
curl http://localhost:8080/api/v1/postal-codes/land/prefix/256
```

**응답 예시** (200 OK):
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "zip_code": "25627",
      "zip_prefix": "256",
      "sido_name": "강원특별자치도",
      "sigungu_name": "강릉시",
      "eupmyeondong_name": "강동면",
      "ri_name": "모전리",
      "is_mountain": false,
      "start_jibun_main": 21
    },
    {
      "id": 2,
      "zip_code": "25628",
      "zip_prefix": "256",
      "sido_name": "강원특별자치도",
      "sigungu_name": "강릉시",
      "eupmyeondong_name": "강동면",
      "ri_name": "산계리",
      "is_mountain": false,
      "start_jibun_main": 1
    }
    // ... 더 많은 결과
  ],
  "total": 856
}
```

---

### 3. 지번주소 복합 검색

**엔드포인트**: `GET /api/v1/postal-codes/land/search`

**목적**: 시도, 시군구, 읍면동, 리명 등 여러 조건으로 유연한 검색

**쿼리 파라미터**:
| 파라미터 | 타입 | 필수 | 설명 | 예시 |
|---------|------|-----|------|------|
| `zip_code` | string | No | 우편번호 (5자리 정확 매칭) | `25627` |
| `zip_prefix` | string | No | 우편번호 앞 3자리 (권장, 빠름) | `256` |
| `sido_name` | string | No | 시도명 (부분 매칭) | `강원` |
| `sigungu_name` | string | No | 시군구명 (부분 매칭) | `강릉` |
| `eupmyeondong_name` | string | No | 읍면동명 (부분 매칭) | `강동면` |
| `ri_name` | string | No | 리명 (부분 매칭) | `모전리` |
| `limit` | int | No | 결과 개수 제한 (기본 100, 최대 1000) | `100` |
| `offset` | int | No | 페이징 오프셋 (기본 0) | `0` |

**사용 시나리오**:

#### 1) 시도명으로 검색
```bash
curl "http://localhost:8080/api/v1/postal-codes/land/search?sido_name=강원&limit=10"
```

#### 2) 복합 조건 검색
```bash
curl "http://localhost:8080/api/v1/postal-codes/land/search?sido_name=강원&eupmyeondong_name=강동면&ri_name=모전리"
```

#### 3) prefix로 빠른 검색 후 필터링
```bash
curl "http://localhost:8080/api/v1/postal-codes/land/search?zip_prefix=256&sigungu_name=강릉"
```

**응답 예시** (200 OK):
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "zip_code": "25627",
      "zip_prefix": "256",
      "sido_name": "강원특별자치도",
      "sigungu_name": "강릉시",
      "eupmyeondong_name": "강동면",
      "ri_name": "모전리",
      "is_mountain": false,
      "start_jibun_main": 21,
      "end_jibun_main": 198
    }
    // ... 더 많은 결과
  ],
  "total": 15
}
```

---

## 📊 응답 형식

### 성공 응답 구조

모든 성공 응답은 다음 형식을 따릅니다:

```typescript
{
  "success": true,        // 항상 true
  "data": [...],          // 결과 배열
  "total": number         // 전체 결과 개수
}
```

### 에러 응답 구조

모든 에러 응답은 다음 형식을 따릅니다:

```typescript
{
  "success": false,       // 항상 false
  "error": string         // 에러 메시지
}
```

### HTTP 상태 코드

| 상태 코드 | 의미 | 예시 |
|----------|------|------|
| 200 OK | 요청 성공 | 검색 결과 반환 |
| 400 Bad Request | 잘못된 요청 | 우편번호 형식 오류 |
| 404 Not Found | 결과 없음 | 우편번호를 찾을 수 없음 |
| 405 Method Not Allowed | 허용되지 않은 HTTP 메서드 | POST 대신 GET 사용 필요 |
| 500 Internal Server Error | 서버 내부 오류 | 데이터베이스 연결 실패 |

---

## ⚡ 성능 최적화

### 검색 성능 비교

31만건 데이터 기준:

| 검색 방법 | 실행시간 | 사용 인덱스 | 권장 |
|-----------|---------|-----------|-----|
| `zip_prefix = '010'` | ~1-5ms | `idx_zip_prefix` | ✅ 최고 성능 |
| `zip_code = '01000'` | ~1-3ms | `idx_zipcode` | ✅ 정확 검색 |
| `zip_code LIKE '010%'` | ~5-15ms | `idx_zipcode` | ⚠️ 비권장 |

### 최적화 팁

#### 1. Prefix 검색 사용
```bash
# ❌ 비권장: LIKE 패턴 사용
curl "http://localhost:8080/api/v1/postal-codes/road/search?zip_code=010"

# ✅ 권장: prefix 엔드포인트 사용
curl "http://localhost:8080/api/v1/postal-codes/road/prefix/010"
```

#### 2. 검색 범위 좁히기
```bash
# ❌ 비효율적: 너무 넓은 범위
curl "http://localhost:8080/api/v1/postal-codes/road/search?sido_name=서울"

# ✅ 효율적: prefix로 범위 좁힌 후 필터링
curl "http://localhost:8080/api/v1/postal-codes/road/search?zip_prefix=010&sido_name=서울"
```

#### 3. 적절한 limit 설정
```bash
# ❌ 비권장: limit 없음 (기본 100)
curl "http://localhost:8080/api/v1/postal-codes/road/search?sido_name=서울"

# ✅ 권장: 필요한 만큼만 조회
curl "http://localhost:8080/api/v1/postal-codes/road/search?sido_name=서울&limit=10"
```

---

## 🔒 보안 고려사항

### 1. SQL Injection 방지

패키지는 GORM을 사용하여 모든 쿼리를 파라미터화하므로 SQL Injection에 안전합니다.

### 2. Rate Limiting

프로덕션 환경에서는 Rate Limiting 미들웨어를 추가하는 것을 권장합니다:

```go
// Gin 예시
import "github.com/gin-contrib/limiter"

r := gin.Default()
r.Use(limiter.RateLimiter(...))
```

### 3. CORS 설정

외부 도메인에서 접근이 필요한 경우 CORS 설정:

```go
// Gin 예시
import "github.com/gin-contrib/cors"

r := gin.Default()
r.Use(cors.Default())
```

---

## 📝 사용 예제

### 프로그래밍 방식 사용

#### Go 클라이언트
```go
package main

import (
    "encoding/json"
    "fmt"
    "net/http"
)

type Response struct {
    Success bool                      `json:"success"`
    Data    []map[string]interface{} `json:"data"`
    Total   int64                     `json:"total"`
}

func main() {
    resp, _ := http.Get("http://localhost:8080/api/v1/postal-codes/road/zipcode/01000")
    defer resp.Body.Close()

    var result Response
    json.NewDecoder(resp.Body).Decode(&result)

    fmt.Printf("Found %d results\n", result.Total)
    for _, item := range result.Data {
        fmt.Printf("주소: %s %s %s\n",
            item["sido_name"],
            item["sigungu_name"],
            item["road_name"])
    }
}
```

#### JavaScript/TypeScript
```javascript
async function searchPostalCode(zipCode) {
  const response = await fetch(
    `http://localhost:8080/api/v1/postal-codes/road/zipcode/${zipCode}`
  );
  const data = await response.json();

  if (data.success) {
    console.log(`Found ${data.total} results`);
    data.data.forEach(item => {
      console.log(`${item.sido_name} ${item.sigungu_name} ${item.road_name}`);
    });
  }
}

searchPostalCode('01000');
```

#### Python
```python
import requests

def search_postal_code(zip_code):
    url = f"http://localhost:8080/api/v1/postal-codes/road/zipcode/{zip_code}"
    response = requests.get(url)
    data = response.json()

    if data['success']:
        print(f"Found {data['total']} results")
        for item in data['data']:
            print(f"{item['sido_name']} {item['sigungu_name']} {item['road_name']}")

search_postal_code('01000')
```

---

## 🧪 테스트

### 자동화된 테스트 실행

패키지는 100개 이상의 자동화된 테스트를 포함합니다:

```bash
# 전체 테스트 실행
go test ./...

# HTTP 핸들러 테스트만 실행
go test ./internal/http

# 커버리지 포함
go test -cover ./...
```

**테스트 커버리지**:
- Repository 계층: CRUD, 검색, 페이징
- Service 계층: 비즈니스 로직, 유효성 검사
- HTTP Handler: 모든 API 엔드포인트
- Integration: 전체 워크플로우

자세한 테스트 정보는 [USAGE.md](./USAGE.md#테스트)를 참조하세요.

### curl을 사용한 수동 테스트

```bash
# 1. 정확한 우편번호 조회
curl -X GET http://localhost:8080/api/v1/postal-codes/road/zipcode/01000 | jq

# 2. prefix 검색
curl -X GET http://localhost:8080/api/v1/postal-codes/road/prefix/010 | jq

# 3. 복합 검색
curl -X GET "http://localhost:8080/api/v1/postal-codes/road/search?sido_name=서울&limit=5" | jq

# 4. 에러 케이스 테스트
curl -X GET http://localhost:8080/api/v1/postal-codes/road/zipcode/invalid | jq
```

### HTTPie를 사용한 테스트

```bash
# HTTPie 설치: brew install httpie

# 1. 정확한 우편번호 조회
http GET http://localhost:8080/api/v1/postal-codes/road/zipcode/01000

# 2. 복합 검색
http GET http://localhost:8080/api/v1/postal-codes/road/search \
  sido_name=="서울" \
  sigungu_name=="강북구" \
  limit==10
```

---

## 📚 관련 문서

- [패키지 README](../README.md) - 패키지 개요 및 설치
- [완전한 사용 가이드](./USAGE.md) - Repository/Service 사용법
- [프로젝트 통합 가이드](./INTEGRATION.md) - 다른 프로젝트에서 사용하기
- [실행 가능한 예제](../examples/README.md) - 코드 예제

---

## 🔧 문제 해결

### Q1. 응답이 없거나 타임아웃 발생

**원인**: 데이터베이스 연결 문제

**해결**:
```bash
# DB 연결 확인
mysql -h HOST -u USER -p DATABASE -e "SELECT COUNT(*) FROM postal_code_roads;"
```

### Q2. 404 Not Found 발생

**원인**: 라우트가 올바르게 등록되지 않음

**해결**:
```go
// 라우트 등록 확인
handler.RegisterRoutes(mux, "/api/v1/postal-codes")  // prefix 확인
```

### Q3. 검색 결과가 너무 많음

**해결**: limit 파라미터 사용
```bash
curl "http://localhost:8080/api/v1/postal-codes/road/search?sido_name=서울&limit=10"
```

---

**Made with ❤️ for Korean Address Management**
