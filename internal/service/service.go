package service

import (
	"fmt"
	"strings"

	postalcode "github.com/oursportsnation/korean-postalcode"
	"github.com/oursportsnation/korean-postalcode/internal/repository"
)

// Service는 우편번호 비즈니스 로직을 제공합니다.
type Service interface {
	// 도로명주소 관련 메서드
	// GetByZipCode는 우편번호로 조회합니다.
	GetByZipCode(zipCode string) ([]postalcode.PostalCodeRoad, error)

	// GetByZipPrefix는 우편번호 앞 3자리로 조회합니다.
	GetByZipPrefix(zipPrefix string, limit, offset int) ([]postalcode.PostalCodeRoad, int64, error)

	// Search는 여러 조건으로 검색합니다.
	Search(params postalcode.SearchParams) ([]postalcode.PostalCodeRoad, int64, error)

	// Upsert는 우편번호 데이터를 생성 또는 업데이트합니다.
	Upsert(road *postalcode.PostalCodeRoad) error

	// BatchUpsert는 여러 우편번호 데이터를 배치로 생성/업데이트합니다.
	BatchUpsert(roads []postalcode.PostalCodeRoad) error

	// ExtractZipPrefix는 우편번호에서 앞 3자리를 추출합니다.
	ExtractZipPrefix(zipCode string) string

	// TruncateRoad는 도로명주소 테이블의 모든 데이터를 삭제합니다.
	TruncateRoad() error

	// 지번주소 관련 메서드
	// GetLandByZipCode는 우편번호로 지번주소를 조회합니다.
	GetLandByZipCode(zipCode string) ([]postalcode.PostalCodeLand, error)

	// GetLandByZipPrefix는 우편번호 앞 3자리로 지번주소를 조회합니다.
	GetLandByZipPrefix(zipPrefix string, limit, offset int) ([]postalcode.PostalCodeLand, int64, error)

	// SearchLand는 여러 조건으로 지번주소를 검색합니다.
	SearchLand(params postalcode.SearchParamsLand) ([]postalcode.PostalCodeLand, int64, error)

	// UpsertLand는 지번주소 데이터를 생성 또는 업데이트합니다.
	UpsertLand(land *postalcode.PostalCodeLand) error

	// BatchUpsertLand는 여러 지번주소 데이터를 배치로 생성/업데이트합니다.
	BatchUpsertLand(lands []postalcode.PostalCodeLand) error

	// TruncateLand는 지번주소 테이블의 모든 데이터를 삭제합니다.
	TruncateLand() error
}

// service는 Service 인터페이스 구현입니다.
type service struct {
	repo repository.Repository
}

// New는 새로운 Service를 생성합니다.
func New(repo repository.Repository) Service {
	return &service{repo: repo}
}

// GetByZipCode는 우편번호로 조회합니다.
func (s *service) GetByZipCode(zipCode string) ([]postalcode.PostalCodeRoad, error) {
	if zipCode == "" {
		return nil, fmt.Errorf("zip code is required")
	}
	if len(zipCode) != 5 {
		return nil, fmt.Errorf("zip code must be 5 digits")
	}
	return s.repo.FindByZipCode(zipCode)
}

// GetByZipPrefix는 우편번호 앞 3자리로 조회합니다.
func (s *service) GetByZipPrefix(zipPrefix string, limit, offset int) ([]postalcode.PostalCodeRoad, int64, error) {
	if zipPrefix == "" {
		return nil, 0, fmt.Errorf("zip prefix is required")
	}
	if len(zipPrefix) != 3 {
		return nil, 0, fmt.Errorf("zip prefix must be 3 digits")
	}

	// 기본값 및 제한 설정
	if limit <= 0 || limit > 100 {
		limit = 10 // 기본 10개
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.FindByZipPrefix(zipPrefix, limit, offset)
}

// Search는 여러 조건으로 검색합니다.
func (s *service) Search(params postalcode.SearchParams) ([]postalcode.PostalCodeRoad, int64, error) {
	// 기본값 설정
	if params.Limit <= 0 || params.Limit > 100 {
		params.Limit = 10
	}
	if params.Page <= 0 {
		params.Page = 1
	}

	return s.repo.Search(params)
}

// Upsert는 우편번호 데이터를 생성 또는 업데이트합니다.
func (s *service) Upsert(road *postalcode.PostalCodeRoad) error {
	// Validation
	if err := s.validate(road); err != nil {
		return err
	}

	// ZipPrefix 자동 설정
	if road.ZipPrefix == "" {
		road.ZipPrefix = s.ExtractZipPrefix(road.ZipCode)
	}

	// If ID is already set, update the existing record directly
	if road.ID > 0 {
		return s.repo.Update(road)
	}

	// 기존 데이터 확인 - UNIQUE index 기준으로 검색
	// UNIQUE index: zip_code, sido_name, sigungu_name, road_name, start_building_main
	existing, err := s.repo.FindByZipCode(road.ZipCode)
	if err != nil {
		return err
	}

	// 정확히 일치하는 레코드 찾기 (UNIQUE constraint 기준)
	for i := range existing {
		if existing[i].SidoName == road.SidoName &&
			existing[i].SigunguName == road.SigunguName &&
			existing[i].RoadName == road.RoadName &&
			existing[i].StartBuildingMain == road.StartBuildingMain {
			// 기존 레코드 업데이트
			road.ID = existing[i].ID
			return s.repo.Update(road)
		}
	}

	// 일치하는 레코드가 없으면 새로 생성
	return s.repo.Create(road)
}

// BatchUpsert는 여러 우편번호 데이터를 배치로 생성/업데이트합니다.
func (s *service) BatchUpsert(roads []postalcode.PostalCodeRoad) error {
	validRoads := make([]postalcode.PostalCodeRoad, 0, len(roads))
	var validationErrors []string

	for i := range roads {
		// Validation
		if err := s.validate(&roads[i]); err != nil {
			// 개별 레코드 실패는 스킵하고 계속 진행
			validationErrors = append(validationErrors, fmt.Sprintf("레코드 %d (우편번호: %s): %v", i, roads[i].ZipCode, err))
			continue
		}

		// ZipPrefix 자동 설정
		if roads[i].ZipPrefix == "" {
			roads[i].ZipPrefix = s.ExtractZipPrefix(roads[i].ZipCode)
		}

		validRoads = append(validRoads, roads[i])
	}

	// Validation 에러가 있으면 출력
	if len(validationErrors) > 0 {
		fmt.Printf("⚠️  Validation 실패: %d개\n", len(validationErrors))
		for i, errMsg := range validationErrors {
			if i < 10 { // 최대 10개만 출력
				fmt.Printf("  - %s\n", errMsg)
			}
		}
		if len(validationErrors) > 10 {
			fmt.Printf("  ... 외 %d개\n", len(validationErrors)-10)
		}
	}

	if len(validRoads) == 0 {
		return fmt.Errorf("no valid records in batch")
	}

	return s.repo.BatchCreate(validRoads)
}

// ExtractZipPrefix는 우편번호에서 앞 3자리를 추출합니다.
func (s *service) ExtractZipPrefix(zipCode string) string {
	zipCode = strings.TrimSpace(zipCode)
	if len(zipCode) >= 3 {
		return zipCode[:3]
	}
	return ""
}

// validate는 우편번호 데이터를 검증합니다.
func (s *service) validate(road *postalcode.PostalCodeRoad) error {
	if road.ZipCode == "" {
		return fmt.Errorf("zip code is required")
	}
	if len(road.ZipCode) != 5 {
		return fmt.Errorf("zip code must be 5 digits")
	}
	if road.SidoName == "" {
		return fmt.Errorf("sido name is required")
	}
	// SigunguName은 선택적 (세종시 등 일부 지역은 시군구가 없음)
	if road.RoadName == "" {
		return fmt.Errorf("road name is required")
	}
	return nil
}

// TruncateRoad는 도로명주소 테이블의 모든 데이터를 삭제합니다.
func (s *service) TruncateRoad() error {
	return s.repo.TruncateRoad()
}

// ============================================================
// 지번주소 관련 메서드
// ============================================================

// GetLandByZipCode는 우편번호로 지번주소를 조회합니다.
func (s *service) GetLandByZipCode(zipCode string) ([]postalcode.PostalCodeLand, error) {
	if zipCode == "" {
		return nil, fmt.Errorf("zip code is required")
	}
	if len(zipCode) != 5 {
		return nil, fmt.Errorf("zip code must be 5 digits")
	}
	return s.repo.FindLandByZipCode(zipCode)
}

// GetLandByZipPrefix는 우편번호 앞 3자리로 지번주소를 조회합니다.
func (s *service) GetLandByZipPrefix(zipPrefix string, limit, offset int) ([]postalcode.PostalCodeLand, int64, error) {
	if zipPrefix == "" {
		return nil, 0, fmt.Errorf("zip prefix is required")
	}
	if len(zipPrefix) != 3 {
		return nil, 0, fmt.Errorf("zip prefix must be 3 digits")
	}

	// 기본값 및 제한 설정
	if limit <= 0 || limit > 100 {
		limit = 10 // 기본 10개
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.FindLandByZipPrefix(zipPrefix, limit, offset)
}

// SearchLand는 여러 조건으로 지번주소를 검색합니다.
func (s *service) SearchLand(params postalcode.SearchParamsLand) ([]postalcode.PostalCodeLand, int64, error) {
	// 기본값 설정
	if params.Limit <= 0 || params.Limit > 100 {
		params.Limit = 10
	}
	if params.Page <= 0 {
		params.Page = 1
	}

	return s.repo.SearchLand(params)
}

// UpsertLand는 지번주소 데이터를 생성 또는 업데이트합니다.
func (s *service) UpsertLand(land *postalcode.PostalCodeLand) error {
	// Validation
	if err := s.validateLand(land); err != nil {
		return err
	}

	// ZipPrefix 자동 설정
	if land.ZipPrefix == "" {
		land.ZipPrefix = s.ExtractZipPrefix(land.ZipCode)
	}

	// If ID is already set, update the existing record directly
	if land.ID > 0 {
		return s.repo.UpdateLand(land)
	}

	// 기존 데이터 확인 - UNIQUE index 기준으로 검색
	// UNIQUE index: zip_code, sido_name, sigungu_name, eupmyeondong_name, ri_name, is_mountain, start_jibun_main
	existing, err := s.repo.FindLandByZipCode(land.ZipCode)
	if err != nil {
		return err
	}

	// 정확히 일치하는 레코드 찾기 (UNIQUE constraint 기준)
	for i := range existing {
		if existing[i].SidoName == land.SidoName &&
			existing[i].SigunguName == land.SigunguName &&
			existing[i].EupmyeondongName == land.EupmyeondongName &&
			existing[i].RiName == land.RiName &&
			existing[i].IsMountain == land.IsMountain &&
			existing[i].StartJibunMain == land.StartJibunMain {
			// 기존 레코드 업데이트
			land.ID = existing[i].ID
			return s.repo.UpdateLand(land)
		}
	}

	// 일치하는 레코드가 없으면 새로 생성
	return s.repo.CreateLand(land)
}

// BatchUpsertLand는 여러 지번주소 데이터를 배치로 생성/업데이트합니다.
func (s *service) BatchUpsertLand(lands []postalcode.PostalCodeLand) error {
	validLands := make([]postalcode.PostalCodeLand, 0, len(lands))
	var validationErrors []string

	for i := range lands {
		// Validation
		if err := s.validateLand(&lands[i]); err != nil {
			// 개별 레코드 실패는 스킵하고 계속 진행
			validationErrors = append(validationErrors, fmt.Sprintf("레코드 %d (우편번호: %s): %v", i, lands[i].ZipCode, err))
			continue
		}

		// ZipPrefix 자동 설정
		if lands[i].ZipPrefix == "" {
			lands[i].ZipPrefix = s.ExtractZipPrefix(lands[i].ZipCode)
		}

		validLands = append(validLands, lands[i])
	}

	// Validation 에러가 있으면 출력
	if len(validationErrors) > 0 {
		fmt.Printf("⚠️  Validation 실패: %d개\n", len(validationErrors))
		for i, errMsg := range validationErrors {
			if i < 10 { // 최대 10개만 출력
				fmt.Printf("  - %s\n", errMsg)
			}
		}
		if len(validationErrors) > 10 {
			fmt.Printf("  ... 외 %d개\n", len(validationErrors)-10)
		}
	}

	if len(validLands) == 0 {
		return fmt.Errorf("no valid records in batch")
	}

	return s.repo.BatchCreateLand(validLands)
}

// validateLand는 지번주소 데이터를 검증합니다.
func (s *service) validateLand(land *postalcode.PostalCodeLand) error {
	if land.ZipCode == "" {
		return fmt.Errorf("zip code is required")
	}
	if len(land.ZipCode) != 5 {
		return fmt.Errorf("zip code must be 5 digits")
	}
	if land.SidoName == "" {
		return fmt.Errorf("sido name is required")
	}
	// SigunguName은 선택적 (세종시 등 일부 지역은 시군구가 없음)
	if land.EupmyeondongName == "" {
		return fmt.Errorf("eupmyeondong name is required")
	}
	return nil
}

// TruncateLand는 지번주소 테이블의 모든 데이터를 삭제합니다.
func (s *service) TruncateLand() error {
	return s.repo.TruncateLand()
}
