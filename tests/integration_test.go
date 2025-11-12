package tests

import (
	"fmt"
	"testing"

	postalcode "github.com/oursportsnation/korean-postalcode"
	"github.com/oursportsnation/korean-postalcode/internal/repository"
	"github.com/oursportsnation/korean-postalcode/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Integration tests verify the full stack works together:
// Repository → Service → Handler (tested through service)

func setupIntegrationTest(t *testing.T) (repository.Repository, service.Service) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&postalcode.PostalCodeRoad{}, &postalcode.PostalCodeLand{})
	require.NoError(t, err)

	repo := repository.New(db)
	svc := service.New(repo)

	return repo, svc
}

// ============================================================
// Road Address Integration Tests
// ============================================================

func TestIntegration_RoadAddress_FullWorkflow(t *testing.T) {
	_, svc := setupIntegrationTest(t)

	// 1. Upsert data (service validates, repository stores)
	road := &postalcode.PostalCodeRoad{
		ZipCode:     "01000",
		SidoName:    "서울특별시",
		SigunguName: "강북구",
		RoadName:    "삼양로177길",
	}

	err := svc.Upsert(road)
	assert.NoError(t, err)
	assert.Equal(t, "010", road.ZipPrefix) // Auto-extracted by service

	// 2. GetByZipCode (service validates, repository queries)
	results, err := svc.GetByZipCode("01000")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "01000", results[0].ZipCode)
	assert.Equal(t, "서울특별시", results[0].SidoName)

	// 3. GetByZipPrefix (service validates, repository queries with pagination)
	results, total, err := svc.GetByZipPrefix("010", 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, results, 1)

	// 4. Search (service validates, repository does complex query)
	searchResults, searchTotal, err := svc.Search(postalcode.SearchParams{
		SidoName: "서울",
		Page:     1,
		Limit:    10,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), searchTotal)
	assert.Len(t, searchResults, 1)

	// 5. Update (upsert with same zipcode updates)
	road.RoadName = "Updated Road"
	err = svc.Upsert(road)
	assert.NoError(t, err)

	// Verify update
	results, err = svc.GetByZipCode("01000")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Road", results[0].RoadName)
}

func TestIntegration_RoadAddress_BatchOperations(t *testing.T) {
	_, svc := setupIntegrationTest(t)

	// Batch upsert
	roads := []postalcode.PostalCodeRoad{
		{ZipCode: "01000", SidoName: "서울특별시", SigunguName: "강북구", RoadName: "삼양로1"},
		{ZipCode: "01001", SidoName: "서울특별시", SigunguName: "강북구", RoadName: "삼양로2"},
		{ZipCode: "01002", SidoName: "서울특별시", SigunguName: "강북구", RoadName: "삼양로3"},
	}

	err := svc.BatchUpsert(roads)
	assert.NoError(t, err)

	// Verify all records
	results, total, err := svc.GetByZipPrefix("010", 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, results, 3)
}

func TestIntegration_RoadAddress_ValidationFlow(t *testing.T) {
	_, svc := setupIntegrationTest(t)

	// Test validation errors flow through the stack
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

func TestIntegration_RoadAddress_ComplexSearch(t *testing.T) {
	_, svc := setupIntegrationTest(t)

	// Seed varied data
	roads := []postalcode.PostalCodeRoad{
		{ZipCode: "01000", SidoName: "서울특별시", SigunguName: "강북구", RoadName: "삼양로1"},
		{ZipCode: "01001", SidoName: "서울특별시", SigunguName: "강북구", RoadName: "삼양로2"},
		{ZipCode: "06000", SidoName: "서울특별시", SigunguName: "강남구", RoadName: "테헤란로"},
		{ZipCode: "21000", SidoName: "부산광역시", SigunguName: "중구", RoadName: "중앙대로"},
	}
	for i := range roads {
		require.NoError(t, svc.Upsert(&roads[i]))
	}

	// Test 1: Search by sido only
	results, total, err := svc.Search(postalcode.SearchParams{
		SidoName: "서울",
		Page:     1,
		Limit:    10,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, results, 3)

	// Test 2: Search by sido + sigungu
	results, total, err = svc.Search(postalcode.SearchParams{
		SidoName:    "서울",
		SigunguName: "강북",
		Page:        1,
		Limit:       10,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, results, 2)

	// Test 3: Search by road name
	results, total, err = svc.Search(postalcode.SearchParams{
		RoadName: "삼양로",
		Page:     1,
		Limit:    10,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, results, 2)
}

func TestIntegration_RoadAddress_Pagination(t *testing.T) {
	_, svc := setupIntegrationTest(t)

	// Create 15 records with different road names to avoid unique constraint
	for i := 0; i < 15; i++ {
		road := &postalcode.PostalCodeRoad{
			ZipCode:           "01000",
			SidoName:          "서울특별시",
			SigunguName:       "강북구",
			RoadName:          fmt.Sprintf("테스트도로%d", i),
			StartBuildingMain: i,
		}
		require.NoError(t, svc.Upsert(road))
	}

	// Test pagination
	// Page 1: limit 10 (should return 10)
	results, total, err := svc.GetByZipPrefix("010", 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(15), total)
	assert.Len(t, results, 10)

	// Page 2: limit 10 offset 10 (should return 5)
	results, total, err = svc.GetByZipPrefix("010", 10, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(15), total)
	assert.Len(t, results, 5)
}

// ============================================================
// Land Address Integration Tests
// ============================================================

func TestIntegration_LandAddress_FullWorkflow(t *testing.T) {
	_, svc := setupIntegrationTest(t)

	// 1. Upsert data
	land := &postalcode.PostalCodeLand{
		ZipCode:          "25627",
		SidoName:         "강원특별자치도",
		SigunguName:      "강릉시",
		EupmyeondongName: "강동면",
		RiName:           "모전리",
	}

	err := svc.UpsertLand(land)
	assert.NoError(t, err)
	assert.Equal(t, "256", land.ZipPrefix)

	// 2. GetLandByZipCode
	results, err := svc.GetLandByZipCode("25627")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "25627", results[0].ZipCode)

	// 3. GetLandByZipPrefix
	results, total, err := svc.GetLandByZipPrefix("256", 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, results, 1)

	// 4. SearchLand
	searchResults, searchTotal, err := svc.SearchLand(postalcode.SearchParamsLand{
		SidoName: "강원",
		Page:     1,
		Limit:    10,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), searchTotal)
	assert.Len(t, searchResults, 1)
}

func TestIntegration_LandAddress_ValidationFlow(t *testing.T) {
	_, svc := setupIntegrationTest(t)

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

func TestIntegration_LandAddress_ComplexSearch(t *testing.T) {
	_, svc := setupIntegrationTest(t)

	// Seed data
	lands := []postalcode.PostalCodeLand{
		{ZipCode: "25627", SidoName: "강원특별자치도", SigunguName: "강릉시", EupmyeondongName: "강동면", RiName: "모전리"},
		{ZipCode: "25628", SidoName: "강원특별자치도", SigunguName: "강릉시", EupmyeondongName: "강동면", RiName: "심곡리"},
		{ZipCode: "25629", SidoName: "강원특별자치도", SigunguName: "강릉시", EupmyeondongName: "왕산면", RiName: "목동리"},
	}
	for i := range lands {
		require.NoError(t, svc.UpsertLand(&lands[i]))
	}

	// Test 1: Search by sido
	_, total, err := svc.SearchLand(postalcode.SearchParamsLand{
		SidoName: "강원",
		Page:     1,
		Limit:    10,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)

	// Test 2: Search by eupmyeondong
	_, total, err = svc.SearchLand(postalcode.SearchParamsLand{
		EupmyeondongName: "강동면",
		Page:             1,
		Limit:            10,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
}

// ============================================================
// Cross-Entity Integration Tests
// ============================================================

func TestIntegration_MixedRoadAndLandAddresses(t *testing.T) {
	_, svc := setupIntegrationTest(t)

	// Insert both road and land addresses
	road := &postalcode.PostalCodeRoad{
		ZipCode:     "01000",
		SidoName:    "서울특별시",
		SigunguName: "강북구",
		RoadName:    "삼양로",
	}
	require.NoError(t, svc.Upsert(road))

	land := &postalcode.PostalCodeLand{
		ZipCode:          "25627",
		SidoName:         "강원특별자치도",
		SigunguName:      "강릉시",
		EupmyeondongName: "강동면",
	}
	require.NoError(t, svc.UpsertLand(land))

	// Verify both can be retrieved independently
	roadResults, err := svc.GetByZipCode("01000")
	assert.NoError(t, err)
	assert.Len(t, roadResults, 1)

	landResults, err := svc.GetLandByZipCode("25627")
	assert.NoError(t, err)
	assert.Len(t, landResults, 1)

	// Verify they don't interfere with each other
	assert.NotEqual(t, roadResults[0].ZipCode, landResults[0].ZipCode)
}

// ============================================================
// Error Propagation Tests
// ============================================================

func TestIntegration_ErrorPropagation(t *testing.T) {
	_, svc := setupIntegrationTest(t)

	// Test that errors propagate correctly from repository → service
	tests := []struct {
		name      string
		operation func() error
		wantErr   bool
	}{
		{
			"get by invalid zipcode",
			func() error {
				_, err := svc.GetByZipCode("123")
				return err
			},
			true,
		},
		{
			"get by invalid prefix",
			func() error {
				_, _, err := svc.GetByZipPrefix("12", 10, 0)
				return err
			},
			true,
		},
		{
			"upsert invalid road",
			func() error {
				return svc.Upsert(&postalcode.PostalCodeRoad{ZipCode: "123"})
			},
			true,
		},
		{
			"upsert invalid land",
			func() error {
				return svc.UpsertLand(&postalcode.PostalCodeLand{ZipCode: "123"})
			},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.operation()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
