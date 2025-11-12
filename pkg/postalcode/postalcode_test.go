package postalcode

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	postalcode "github.com/oursportsnation/korean-postalcode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Public API tests verify that the exported package API works correctly
// and provides a clean interface for external users.

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&postalcode.PostalCodeRoad{}, &postalcode.PostalCodeLand{})
	require.NoError(t, err)

	return db
}

// ============================================================
// Factory Function Tests
// ============================================================

func TestPublicAPI_NewRepository(t *testing.T) {
	db := setupTestDB(t)

	// Test that NewRepository returns a working repository
	repo := NewRepository(db)
	assert.NotNil(t, repo)

	// Verify it can perform basic operations
	road := &postalcode.PostalCodeRoad{
		ZipCode:     "01000",
		ZipPrefix:   "010",
		SidoName:    "서울특별시",
		SigunguName: "강북구",
		RoadName:    "삼양로",
	}
	err := repo.Create(road)
	assert.NoError(t, err)
	assert.NotZero(t, road.ID)
}

func TestPublicAPI_NewService(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)

	// Test that NewService returns a working service
	svc := NewService(repo)
	assert.NotNil(t, svc)

	// Verify it can perform operations
	road := &postalcode.PostalCodeRoad{
		ZipCode:     "01000",
		SidoName:    "서울특별시",
		SigunguName: "강북구",
		RoadName:    "삼양로",
	}
	err := svc.Upsert(road)
	assert.NoError(t, err)
	assert.Equal(t, "010", road.ZipPrefix) // Auto-extracted
}

func TestPublicAPI_NewImporter(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	svc := NewService(repo)

	// Test that NewImporter returns a working importer
	imp := NewImporter(svc)
	assert.NotNil(t, imp)

	// Verify it has the expected interface
	// (Just checking it's not nil and has the right type)
	_, ok := interface{}(imp).(Importer)
	assert.True(t, ok)
}

// ============================================================
// HTTP Routes Registration Tests
// ============================================================

func TestPublicAPI_RegisterHTTPRoutes(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	svc := NewService(repo)

	// Seed data
	road := &postalcode.PostalCodeRoad{
		ZipCode:     "01000",
		SidoName:    "서울특별시",
		SigunguName: "강북구",
		RoadName:    "삼양로",
	}
	require.NoError(t, svc.Upsert(road))

	// Test route registration
	mux := http.NewServeMux()
	RegisterHTTPRoutes(svc, mux, "/api/v1/postal-codes")

	// Test registered route works
	req := httptest.NewRequest("GET", "/api/v1/postal-codes/road/zipcode/01000", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
}

func TestPublicAPI_RegisterHTTPRoutes_MultipleEndpoints(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	svc := NewService(repo)

	// Seed data
	road := &postalcode.PostalCodeRoad{
		ZipCode:     "01000",
		SidoName:    "서울특별시",
		SigunguName: "강북구",
		RoadName:    "삼양로",
	}
	require.NoError(t, svc.Upsert(road))

	mux := http.NewServeMux()
	RegisterHTTPRoutes(svc, mux, "/api/v1/postal-codes")

	// Test all endpoints are registered
	tests := []struct {
		name string
		path string
	}{
		{"road zipcode", "/api/v1/postal-codes/road/zipcode/01000"},
		{"road prefix", "/api/v1/postal-codes/road/prefix/010"},
		{"road search", "/api/v1/postal-codes/road/search?sido_name=서울"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

func TestPublicAPI_RegisterGinRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDB(t)
	repo := NewRepository(db)
	svc := NewService(repo)

	// Seed data
	road := &postalcode.PostalCodeRoad{
		ZipCode:     "01000",
		SidoName:    "서울특별시",
		SigunguName: "강북구",
		RoadName:    "삼양로",
	}
	require.NoError(t, svc.Upsert(road))

	// Test Gin route registration
	router := gin.New()
	rg := router.Group("/api/v1/postal-codes")
	RegisterGinRoutes(svc, rg)

	// Test registered route works
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/postal-codes/road/zipcode/01000", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
}

func TestPublicAPI_RegisterGinRoutes_MultipleEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	db := setupTestDB(t)
	repo := NewRepository(db)
	svc := NewService(repo)

	// Seed data
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

	router := gin.New()
	rg := router.Group("/api/v1/postal-codes")
	RegisterGinRoutes(svc, rg)

	// Test all endpoints are registered
	tests := []struct {
		name string
		path string
	}{
		{"road zipcode", "/api/v1/postal-codes/road/zipcode/01000"},
		{"road prefix", "/api/v1/postal-codes/road/prefix/010"},
		{"road search", "/api/v1/postal-codes/road/search?sido_name=서울"},
		{"land zipcode", "/api/v1/postal-codes/land/zipcode/25627"},
		{"land prefix", "/api/v1/postal-codes/land/prefix/256"},
		{"land search", "/api/v1/postal-codes/land/search?sido_name=강원"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

// ============================================================
// End-to-End Public API Tests
// ============================================================

func TestPublicAPI_EndToEnd_StandardWorkflow(t *testing.T) {
	// This test simulates how an external user would use the public API

	// 1. Setup database
	db := setupTestDB(t)

	// 2. Create repository using public API
	repo := NewRepository(db)
	assert.NotNil(t, repo)

	// 3. Create service using public API
	svc := NewService(repo)
	assert.NotNil(t, svc)

	// 4. Insert data
	road := &postalcode.PostalCodeRoad{
		ZipCode:     "01000",
		SidoName:    "서울특별시",
		SigunguName: "강북구",
		RoadName:    "삼양로",
	}
	err := svc.Upsert(road)
	assert.NoError(t, err)

	// 5. Query data
	results, err := svc.GetByZipCode("01000")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "서울특별시", results[0].SidoName)

	// 6. Setup HTTP handler
	mux := http.NewServeMux()
	RegisterHTTPRoutes(svc, mux, "/api")

	// 7. Test HTTP endpoint
	req := httptest.NewRequest("GET", "/api/road/zipcode/01000", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPublicAPI_EndToEnd_WithImporter(t *testing.T) {
	// Test the complete workflow including importer

	// 1. Setup
	db := setupTestDB(t)
	repo := NewRepository(db)
	svc := NewService(repo)
	imp := NewImporter(svc)

	assert.NotNil(t, repo)
	assert.NotNil(t, svc)
	assert.NotNil(t, imp)

	// 2. Import would normally be from file, but we'll use direct upsert
	roads := []postalcode.PostalCodeRoad{
		{ZipCode: "01000", SidoName: "서울특별시", SigunguName: "강북구", RoadName: "삼양로1"},
		{ZipCode: "01001", SidoName: "서울특별시", SigunguName: "강북구", RoadName: "삼양로2"},
	}
	err := svc.BatchUpsert(roads)
	assert.NoError(t, err)

	// 3. Verify data
	results, total, err := svc.GetByZipPrefix("010", 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, results, 2)
}

func TestPublicAPI_EndToEnd_GinWorkflow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Complete Gin workflow

	// 1. Setup
	db := setupTestDB(t)
	repo := NewRepository(db)
	svc := NewService(repo)

	// 2. Seed data
	road := &postalcode.PostalCodeRoad{
		ZipCode:     "01000",
		SidoName:    "서울특별시",
		SigunguName: "강북구",
		RoadName:    "삼양로",
	}
	require.NoError(t, svc.Upsert(road))

	// 3. Setup Gin router
	router := gin.New()
	RegisterGinRoutes(svc, router.Group("/api/v1/postal-codes"))

	// 4. Test endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/postal-codes/road/zipcode/01000", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	assert.NotNil(t, resp["data"])
}

// ============================================================
// Interface Compatibility Tests
// ============================================================

func TestPublicAPI_InterfaceCompatibility(t *testing.T) {
	// Verify that public interfaces are compatible with internal implementations

	db := setupTestDB(t)
	repo := NewRepository(db)
	svc := NewService(repo)
	imp := NewImporter(svc)

	// Test Repository interface
	_, ok := interface{}(repo).(Repository)
	assert.True(t, ok, "Repository should implement Repository interface")

	// Test Service interface
	_, ok = interface{}(svc).(Service)
	assert.True(t, ok, "Service should implement Service interface")

	// Test Importer interface
	_, ok = interface{}(imp).(Importer)
	assert.True(t, ok, "Importer should implement Importer interface")
}

func TestPublicAPI_TypeAliases(t *testing.T) {
	// Verify that type aliases work correctly

	db := setupTestDB(t)

	// These should compile without errors
	var repo Repository = NewRepository(db)
	var svc Service = NewService(repo)
	var imp Importer = NewImporter(svc)

	assert.NotNil(t, repo)
	assert.NotNil(t, svc)
	assert.NotNil(t, imp)
}
