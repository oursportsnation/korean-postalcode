package service

import (
	"fmt"
	"testing"

	postalcode "github.com/oursportsnation/korean-postalcode"
	"github.com/oursportsnation/korean-postalcode/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestService(t *testing.T) Service {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&postalcode.PostalCodeRoad{}, &postalcode.PostalCodeLand{})
	require.NoError(t, err)

	repo := repository.New(db)
	return New(repo)
}

// ============================================================
// Road Address Service Tests
// ============================================================

func TestService_GetByZipCode_Success(t *testing.T) {
	svc := setupTestService(t)

	// Create test data
	road := &postalcode.PostalCodeRoad{
		ZipCode:     "01000",
		ZipPrefix:   "010",
		SidoName:    "서울특별시",
		SigunguName: "강북구",
		RoadName:    "삼양로177길",
	}
	require.NoError(t, svc.Upsert(road))

	// Test
	results, err := svc.GetByZipCode("01000")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "01000", results[0].ZipCode)
}

func TestService_GetByZipCode_Validation(t *testing.T) {
	svc := setupTestService(t)

	tests := []struct {
		name    string
		zipCode string
		wantErr bool
	}{
		{"empty zipcode", "", true},
		{"invalid length", "123", true},
		{"valid zipcode", "01000", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.GetByZipCode(tt.zipCode)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_GetByZipPrefix_Success(t *testing.T) {
	svc := setupTestService(t)

	// Create test data
	roads := []postalcode.PostalCodeRoad{
		{ZipCode: "01000", ZipPrefix: "010", SidoName: "서울특별시", SigunguName: "강북구", RoadName: "삼양로1"},
		{ZipCode: "01001", ZipPrefix: "010", SidoName: "서울특별시", SigunguName: "강북구", RoadName: "삼양로2"},
	}
	for i := range roads {
		require.NoError(t, svc.Upsert(&roads[i]))
	}

	// Test
	results, total, err := svc.GetByZipPrefix("010", 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, results, 2)
}

func TestService_GetByZipPrefix_Validation(t *testing.T) {
	svc := setupTestService(t)

	tests := []struct {
		name      string
		zipPrefix string
		wantErr   bool
	}{
		{"empty prefix", "", true},
		{"invalid length", "12", true},
		{"valid prefix", "010", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := svc.GetByZipPrefix(tt.zipPrefix, 10, 0)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_GetByZipPrefix_Pagination(t *testing.T) {
	svc := setupTestService(t)

	// Create test data (15 records)
	for i := 0; i < 15; i++ {
		road := &postalcode.PostalCodeRoad{
			ZipCode:           "01000",
			ZipPrefix:         "010",
			SidoName:          "서울특별시",
			SigunguName:       "강북구",
			RoadName:          fmt.Sprintf("테스트도로%d", i),
			StartBuildingMain: i,
		}
		require.NoError(t, svc.Upsert(road))
	}

	// Test: Default limit (10)
	results, total, err := svc.GetByZipPrefix("010", 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(15), total)
	assert.Len(t, results, 10) // Default limit

	// Test: Custom limit
	results, total, err = svc.GetByZipPrefix("010", 5, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(15), total)
	assert.Len(t, results, 5)
}

func TestService_Search_Success(t *testing.T) {
	svc := setupTestService(t)

	// Create test data
	roads := []postalcode.PostalCodeRoad{
		{ZipCode: "01000", ZipPrefix: "010", SidoName: "서울특별시", SigunguName: "강북구", RoadName: "삼양로"},
		{ZipCode: "06000", ZipPrefix: "060", SidoName: "서울특별시", SigunguName: "강남구", RoadName: "테헤란로"},
	}
	for i := range roads {
		require.NoError(t, svc.Upsert(&roads[i]))
	}

	// Test: Search by SidoName
	params := postalcode.SearchParams{
		SidoName: "서울",
		Page:     1,
		Limit:    10,
	}
	results, total, err := svc.Search(params)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, results, 2)
}

func TestService_Upsert_AutoZipPrefix(t *testing.T) {
	svc := setupTestService(t)

	road := &postalcode.PostalCodeRoad{
		ZipCode:     "01000",
		SidoName:    "서울특별시",
		SigunguName: "강북구",
		RoadName:    "삼양로177길",
	}

	err := svc.Upsert(road)
	assert.NoError(t, err)
	assert.Equal(t, "010", road.ZipPrefix) // Auto-extracted
}

func TestService_Upsert_Validation(t *testing.T) {
	svc := setupTestService(t)

	tests := []struct {
		name    string
		road    *postalcode.PostalCodeRoad
		wantErr bool
	}{
		{
			"missing zipcode",
			&postalcode.PostalCodeRoad{SidoName: "서울", SigunguName: "강북구", RoadName: "삼양로"},
			true,
		},
		{
			"invalid zipcode length",
			&postalcode.PostalCodeRoad{ZipCode: "123", SidoName: "서울", SigunguName: "강북구", RoadName: "삼양로"},
			true,
		},
		{
			"missing sido name",
			&postalcode.PostalCodeRoad{ZipCode: "01000", SigunguName: "강북구", RoadName: "삼양로"},
			true,
		},
		{
			"missing road name",
			&postalcode.PostalCodeRoad{ZipCode: "01000", SidoName: "서울", SigunguName: "강북구"},
			true,
		},
		{
			"valid road",
			&postalcode.PostalCodeRoad{ZipCode: "01000", SidoName: "서울특별시", SigunguName: "강북구", RoadName: "삼양로"},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Upsert(tt.road)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_BatchUpsert_Success(t *testing.T) {
	svc := setupTestService(t)

	roads := []postalcode.PostalCodeRoad{
		{ZipCode: "01000", SidoName: "서울특별시", SigunguName: "강북구", RoadName: "삼양로1"},
		{ZipCode: "01001", SidoName: "서울특별시", SigunguName: "강북구", RoadName: "삼양로2"},
	}

	err := svc.BatchUpsert(roads)
	assert.NoError(t, err)

	// Verify
	results, total, err := svc.GetByZipPrefix("010", 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, results, 2)
}

func TestService_BatchUpsert_PartialFailure(t *testing.T) {
	svc := setupTestService(t)

	roads := []postalcode.PostalCodeRoad{
		{ZipCode: "01000", SidoName: "서울특별시", SigunguName: "강북구", RoadName: "삼양로1"}, // Valid
		{ZipCode: "", SidoName: "서울특별시", SigunguName: "강북구", RoadName: "삼양로2"},      // Invalid
		{ZipCode: "01002", SidoName: "서울특별시", SigunguName: "강북구", RoadName: "삼양로3"}, // Valid
	}

	err := svc.BatchUpsert(roads)
	assert.NoError(t, err) // Should continue despite individual failures

	// Verify: Only valid records inserted
	_, total, err := svc.GetByZipPrefix("010", 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total) // Only 2 valid records
}

func TestService_ExtractZipPrefix(t *testing.T) {
	svc := setupTestService(t)

	tests := []struct {
		zipCode string
		want    string
	}{
		{"01000", "010"},
		{"99999", "999"},
		{"12", ""},
		{"", ""},
		{"  01000  ", "010"}, // Trim whitespace
	}

	for _, tt := range tests {
		t.Run(tt.zipCode, func(t *testing.T) {
			result := svc.ExtractZipPrefix(tt.zipCode)
			assert.Equal(t, tt.want, result)
		})
	}
}

// ============================================================
// Land Address Service Tests
// ============================================================

func TestService_GetLandByZipCode_Success(t *testing.T) {
	svc := setupTestService(t)

	// Create test data
	land := &postalcode.PostalCodeLand{
		ZipCode:          "25627",
		ZipPrefix:        "256",
		SidoName:         "강원특별자치도",
		SigunguName:      "강릉시",
		EupmyeondongName: "강동면",
	}
	require.NoError(t, svc.UpsertLand(land))

	// Test
	results, err := svc.GetLandByZipCode("25627")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "25627", results[0].ZipCode)
}

func TestService_SearchLand_Success(t *testing.T) {
	svc := setupTestService(t)

	// Create test data
	lands := []postalcode.PostalCodeLand{
		{ZipCode: "25627", ZipPrefix: "256", SidoName: "강원특별자치도", SigunguName: "강릉시", EupmyeondongName: "강동면"},
		{ZipCode: "25628", ZipPrefix: "256", SidoName: "강원특별자치도", SigunguName: "강릉시", EupmyeondongName: "강동면"},
	}
	for i := range lands {
		require.NoError(t, svc.UpsertLand(&lands[i]))
	}

	// Test
	params := postalcode.SearchParamsLand{
		SidoName: "강원",
		Page:     1,
		Limit:    10,
	}
	results, total, err := svc.SearchLand(params)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, results, 2)
}

func TestService_UpsertLand_Validation(t *testing.T) {
	svc := setupTestService(t)

	tests := []struct {
		name    string
		land    *postalcode.PostalCodeLand
		wantErr bool
	}{
		{
			"missing zipcode",
			&postalcode.PostalCodeLand{SidoName: "강원", SigunguName: "강릉시", EupmyeondongName: "강동면"},
			true,
		},
		{
			"invalid zipcode length",
			&postalcode.PostalCodeLand{ZipCode: "123", SidoName: "강원", SigunguName: "강릉시", EupmyeondongName: "강동면"},
			true,
		},
		{
			"missing eupmyeondong name",
			&postalcode.PostalCodeLand{ZipCode: "25627", SidoName: "강원", SigunguName: "강릉시"},
			true,
		},
		{
			"valid land",
			&postalcode.PostalCodeLand{ZipCode: "25627", SidoName: "강원특별자치도", SigunguName: "강릉시", EupmyeondongName: "강동면"},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.UpsertLand(tt.land)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
