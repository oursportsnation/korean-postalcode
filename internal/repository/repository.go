package repository

import (
	postalcode "github.com/oursportsnation/korean-postalcode"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository는 우편번호 데이터 접근 인터페이스입니다.
type Repository interface {
	// 도로명주소 관련 메서드
	// FindByZipCode는 우편번호로 조회합니다.
	FindByZipCode(zipCode string) ([]postalcode.PostalCodeRoad, error)

	// FindByZipPrefix는 우편번호 앞 3자리로 조회합니다.
	FindByZipPrefix(zipPrefix string, limit, offset int) ([]postalcode.PostalCodeRoad, int64, error)

	// Search는 여러 조건으로 검색합니다.
	Search(params postalcode.SearchParams) ([]postalcode.PostalCodeRoad, int64, error)

	// Create는 새로운 우편번호 데이터를 생성합니다.
	Create(road *postalcode.PostalCodeRoad) error

	// BatchCreate는 여러 우편번호 데이터를 배치로 생성합니다.
	BatchCreate(roads []postalcode.PostalCodeRoad) error

	// Update는 우편번호 데이터를 업데이트합니다.
	Update(road *postalcode.PostalCodeRoad) error

	// Delete는 우편번호 데이터를 삭제합니다.
	Delete(id uint) error

	// TruncateRoad는 도로명주소 테이블의 모든 데이터를 삭제합니다.
	TruncateRoad() error

	// 지번주소 관련 메서드
	// FindLandByZipCode는 우편번호로 지번주소를 조회합니다.
	FindLandByZipCode(zipCode string) ([]postalcode.PostalCodeLand, error)

	// FindLandByZipPrefix는 우편번호 앞 3자리로 지번주소를 조회합니다.
	FindLandByZipPrefix(zipPrefix string, limit, offset int) ([]postalcode.PostalCodeLand, int64, error)

	// SearchLand는 여러 조건으로 지번주소를 검색합니다.
	SearchLand(params postalcode.SearchParamsLand) ([]postalcode.PostalCodeLand, int64, error)

	// CreateLand는 새로운 지번주소 데이터를 생성합니다.
	CreateLand(land *postalcode.PostalCodeLand) error

	// BatchCreateLand는 여러 지번주소 데이터를 배치로 생성합니다.
	BatchCreateLand(lands []postalcode.PostalCodeLand) error

	// UpdateLand는 지번주소 데이터를 업데이트합니다.
	UpdateLand(land *postalcode.PostalCodeLand) error

	// DeleteLand는 지번주소 데이터를 삭제합니다.
	DeleteLand(id uint) error

	// TruncateLand는 지번주소 테이블의 모든 데이터를 삭제합니다.
	TruncateLand() error
}

// gormRepository는 GORM 기반 Repository 구현입니다.
type gormRepository struct {
	db *gorm.DB
}

// New는 새로운 Repository를 생성합니다.
func New(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

// FindByZipCode는 우편번호로 조회합니다.
func (r *gormRepository) FindByZipCode(zipCode string) ([]postalcode.PostalCodeRoad, error) {
	var roads []postalcode.PostalCodeRoad
	err := r.db.Where("zip_code = ?", zipCode).Find(&roads).Error
	return roads, err
}

// FindByZipPrefix는 우편번호 앞 3자리로 조회합니다.
func (r *gormRepository) FindByZipPrefix(zipPrefix string, limit, offset int) ([]postalcode.PostalCodeRoad, int64, error) {
	var roads []postalcode.PostalCodeRoad
	var total int64

	query := r.db.Model(&postalcode.PostalCodeRoad{}).Where("zip_prefix = ?", zipPrefix)

	// 총 개수 조회
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 페이징 적용
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	// 조회
	err := query.Find(&roads).Error
	return roads, total, err
}

// Search는 여러 조건으로 검색합니다.
func (r *gormRepository) Search(params postalcode.SearchParams) ([]postalcode.PostalCodeRoad, int64, error) {
	var roads []postalcode.PostalCodeRoad
	var total int64

	query := r.db.Model(&postalcode.PostalCodeRoad{})

	// 조건 추가
	if params.ZipCode != "" {
		query = query.Where("zip_code = ?", params.ZipCode)
	}
	if params.ZipPrefix != "" {
		query = query.Where("zip_prefix = ?", params.ZipPrefix)
	}
	if params.SidoName != "" {
		query = query.Where("sido_name LIKE ?", "%"+params.SidoName+"%")
	}
	if params.SigunguName != "" {
		query = query.Where("sigungu_name LIKE ?", "%"+params.SigunguName+"%")
	}
	if params.RoadName != "" {
		query = query.Where("road_name LIKE ?", "%"+params.RoadName+"%")
	}

	// 총 개수 조회
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 페이징 (page를 offset으로 변환)
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	} else {
		query = query.Limit(10) // 기본 10개
	}

	// page 기반 offset 계산
	offset := (params.Page - 1) * params.Limit
	if offset > 0 {
		query = query.Offset(offset)
	}

	// 조회
	err := query.Find(&roads).Error
	return roads, total, err
}

// Create는 새로운 우편번호 데이터를 생성합니다.
func (r *gormRepository) Create(road *postalcode.PostalCodeRoad) error {
	return r.db.Create(road).Error
}

// BatchCreate는 여러 우편번호 데이터를 배치로 생성합니다.
func (r *gormRepository) BatchCreate(roads []postalcode.PostalCodeRoad) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "zip_code"}, {Name: "sido_name"}, {Name: "sigungu_name"}, {Name: "road_name"}, {Name: "start_building_main"}},
		UpdateAll: true,
	}).Create(&roads).Error
}

// Update는 우편번호 데이터를 업데이트합니다.
func (r *gormRepository) Update(road *postalcode.PostalCodeRoad) error {
	return r.db.Save(road).Error
}

// Delete는 우편번호 데이터를 삭제합니다.
func (r *gormRepository) Delete(id uint) error {
	return r.db.Delete(&postalcode.PostalCodeRoad{}, id).Error
}

// TruncateRoad는 도로명주소 테이블의 모든 데이터를 삭제합니다.
func (r *gormRepository) TruncateRoad() error {
	// MySQL과 SQLite 모두 지원
	// MySQL의 경우 TRUNCATE가 빠르지만, SQLite는 DELETE를 사용
	dialect := r.db.Dialector.Name()

	if dialect == "mysql" {
		return r.db.Exec("TRUNCATE TABLE postal_code_roads").Error
	}

	// SQLite 또는 다른 DB의 경우
	// 1. 모든 데이터 삭제
	if err := r.db.Exec("DELETE FROM postal_code_roads").Error; err != nil {
		return err
	}

	// 2. AUTO_INCREMENT 리셋 (SQLite의 경우)
	if dialect == "sqlite" {
		return r.db.Exec("DELETE FROM sqlite_sequence WHERE name='postal_code_roads'").Error
	}

	return nil
}

// ============================================================
// 지번주소 관련 메서드
// ============================================================

// FindLandByZipCode는 우편번호로 지번주소를 조회합니다.
func (r *gormRepository) FindLandByZipCode(zipCode string) ([]postalcode.PostalCodeLand, error) {
	var lands []postalcode.PostalCodeLand
	err := r.db.Where("zip_code = ?", zipCode).Find(&lands).Error
	return lands, err
}

// FindLandByZipPrefix는 우편번호 앞 3자리로 지번주소를 조회합니다.
func (r *gormRepository) FindLandByZipPrefix(zipPrefix string, limit, offset int) ([]postalcode.PostalCodeLand, int64, error) {
	var lands []postalcode.PostalCodeLand
	var total int64

	query := r.db.Model(&postalcode.PostalCodeLand{}).Where("zip_prefix = ?", zipPrefix)

	// 총 개수 조회
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 페이징 적용
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	// 조회
	err := query.Find(&lands).Error
	return lands, total, err
}

// SearchLand는 여러 조건으로 지번주소를 검색합니다.
func (r *gormRepository) SearchLand(params postalcode.SearchParamsLand) ([]postalcode.PostalCodeLand, int64, error) {
	var lands []postalcode.PostalCodeLand
	var total int64

	query := r.db.Model(&postalcode.PostalCodeLand{})

	// 조건 추가
	if params.ZipCode != "" {
		query = query.Where("zip_code = ?", params.ZipCode)
	}
	if params.ZipPrefix != "" {
		query = query.Where("zip_prefix = ?", params.ZipPrefix)
	}
	if params.SidoName != "" {
		query = query.Where("sido_name LIKE ?", "%"+params.SidoName+"%")
	}
	if params.SigunguName != "" {
		query = query.Where("sigungu_name LIKE ?", "%"+params.SigunguName+"%")
	}
	if params.EupmyeondongName != "" {
		query = query.Where("eupmyeondong_name LIKE ?", "%"+params.EupmyeondongName+"%")
	}
	if params.RiName != "" {
		query = query.Where("ri_name LIKE ?", "%"+params.RiName+"%")
	}

	// 총 개수 조회
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 페이징 (page를 offset으로 변환)
	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	} else {
		query = query.Limit(10) // 기본 10개
	}

	// page 기반 offset 계산
	offset := (params.Page - 1) * params.Limit
	if offset > 0 {
		query = query.Offset(offset)
	}

	// 조회
	err := query.Find(&lands).Error
	return lands, total, err
}

// CreateLand는 새로운 지번주소 데이터를 생성합니다.
func (r *gormRepository) CreateLand(land *postalcode.PostalCodeLand) error {
	return r.db.Create(land).Error
}

// BatchCreateLand는 여러 지번주소 데이터를 배치로 생성합니다.
func (r *gormRepository) BatchCreateLand(lands []postalcode.PostalCodeLand) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "zip_code"}, {Name: "sido_name"}, {Name: "sigungu_name"}, {Name: "eupmyeondong_name"}, {Name: "ri_name"}, {Name: "is_mountain"}, {Name: "start_jibun_main"}},
		UpdateAll: true,
	}).Create(&lands).Error
}

// UpdateLand는 지번주소 데이터를 업데이트합니다.
func (r *gormRepository) UpdateLand(land *postalcode.PostalCodeLand) error {
	return r.db.Save(land).Error
}

// DeleteLand는 지번주소 데이터를 삭제합니다.
func (r *gormRepository) DeleteLand(id uint) error {
	return r.db.Delete(&postalcode.PostalCodeLand{}, id).Error
}

// TruncateLand는 지번주소 테이블의 모든 데이터를 삭제합니다.
func (r *gormRepository) TruncateLand() error {
	// MySQL과 SQLite 모두 지원
	// MySQL의 경우 TRUNCATE가 빠르지만, SQLite는 DELETE를 사용
	dialect := r.db.Dialector.Name()

	if dialect == "mysql" {
		return r.db.Exec("TRUNCATE TABLE postal_code_lands").Error
	}

	// SQLite 또는 다른 DB의 경우
	// 1. 모든 데이터 삭제
	if err := r.db.Exec("DELETE FROM postal_code_lands").Error; err != nil {
		return err
	}

	// 2. AUTO_INCREMENT 리셋 (SQLite의 경우)
	if dialect == "sqlite" {
		return r.db.Exec("DELETE FROM sqlite_sequence WHERE name='postal_code_lands'").Error
	}

	return nil
}
