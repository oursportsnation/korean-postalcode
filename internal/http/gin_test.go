package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	postalcode "github.com/oursportsnation/korean-postalcode"
	"github.com/oursportsnation/korean-postalcode/internal/repository"
	"github.com/oursportsnation/korean-postalcode/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestGinHandler(t *testing.T) (*GinHandler, *gin.Engine) {
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&postalcode.PostalCodeRoad{}, &postalcode.PostalCodeLand{})
	require.NoError(t, err)

	repo := repository.New(db)
	svc := service.New(repo)
	handler := NewGin(svc)

	router := gin.New()
	rg := router.Group("/api/v1/postal-codes")
	handler.RegisterGinRoutes(rg)

	return handler, router
}

func seedGinTestData(t *testing.T, handler *GinHandler) {
	// Seed road address data
	roads := []postalcode.PostalCodeRoad{
		{ZipCode: "01000", ZipPrefix: "010", SidoName: "서울특별시", SigunguName: "강북구", RoadName: "삼양로1"},
		{ZipCode: "01001", ZipPrefix: "010", SidoName: "서울특별시", SigunguName: "강북구", RoadName: "삼양로2"},
		{ZipCode: "06000", ZipPrefix: "060", SidoName: "서울특별시", SigunguName: "강남구", RoadName: "테헤란로"},
	}
	for i := range roads {
		require.NoError(t, handler.service.Upsert(&roads[i]))
	}

	// Seed land address data
	lands := []postalcode.PostalCodeLand{
		{ZipCode: "25627", ZipPrefix: "256", SidoName: "강원특별자치도", SigunguName: "강릉시", EupmyeondongName: "강동면", RiName: "모전리"},
		{ZipCode: "25628", ZipPrefix: "256", SidoName: "강원특별자치도", SigunguName: "강릉시", EupmyeondongName: "강동면", RiName: "심곡리"},
	}
	for i := range lands {
		require.NoError(t, handler.service.UpsertLand(&lands[i]))
	}
}

// ============================================================
// Road Address Gin Handler Tests
// ============================================================

func TestGinHandler_GetByZipCode_Success(t *testing.T) {
	handler, router := setupTestGinHandler(t)
	seedGinTestData(t, handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/postal-codes/road/zipcode/01000", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	assert.NotNil(t, resp["data"])
	assert.Equal(t, float64(1), resp["total"].(float64))
}

func TestGinHandler_GetByZipCode_NotFound(t *testing.T) {
	handler, router := setupTestGinHandler(t)
	seedGinTestData(t, handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/postal-codes/road/zipcode/99999", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp["success"].(bool))
	assert.Contains(t, resp["error"].(string), "not found")
}

func TestGinHandler_GetByZipCode_InvalidZipCode(t *testing.T) {
	handler, router := setupTestGinHandler(t)
	seedGinTestData(t, handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/postal-codes/road/zipcode/123", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp["success"].(bool))
	assert.NotEmpty(t, resp["error"])
}

func TestGinHandler_GetByZipPrefix_Success(t *testing.T) {
	handler, router := setupTestGinHandler(t)
	seedGinTestData(t, handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/postal-codes/road/prefix/010?page=1&limit=10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	assert.NotNil(t, resp["data"])
	assert.Equal(t, float64(2), resp["total"].(float64))
}

func TestGinHandler_GetByZipPrefix_Pagination(t *testing.T) {
	handler, router := setupTestGinHandler(t)
	seedGinTestData(t, handler)

	// Test page 1 with limit 1
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/postal-codes/road/prefix/010?page=1&limit=1", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	assert.Equal(t, float64(2), resp["total"].(float64))

	// Verify data is a slice with 1 item
	data, ok := resp["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 1)
}

func TestGinHandler_GetByZipPrefix_InvalidPrefix(t *testing.T) {
	handler, router := setupTestGinHandler(t)
	seedGinTestData(t, handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/postal-codes/road/prefix/12", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp["success"].(bool))
	assert.NotEmpty(t, resp["error"])
}

func TestGinHandler_Search_Success(t *testing.T) {
	handler, router := setupTestGinHandler(t)
	seedGinTestData(t, handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/postal-codes/road/search?sido_name=서울&page=1&limit=10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	assert.NotNil(t, resp["data"])
	assert.Equal(t, float64(3), resp["total"].(float64))
}

func TestGinHandler_Search_MultipleParams(t *testing.T) {
	handler, router := setupTestGinHandler(t)
	seedGinTestData(t, handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/postal-codes/road/search?sido_name=서울&sigungu_name=강북", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	assert.Equal(t, float64(2), resp["total"].(float64))
}

func TestGinHandler_Search_NoResults(t *testing.T) {
	handler, router := setupTestGinHandler(t)
	seedGinTestData(t, handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/postal-codes/road/search?sido_name=부산", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	assert.Equal(t, float64(0), resp["total"].(float64))
}

// ============================================================
// Land Address Gin Handler Tests
// ============================================================

func TestGinHandler_GetLandByZipCode_Success(t *testing.T) {
	handler, router := setupTestGinHandler(t)
	seedGinTestData(t, handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/postal-codes/land/zipcode/25627", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	assert.NotNil(t, resp["data"])
	assert.Equal(t, float64(1), resp["total"].(float64))
}

func TestGinHandler_GetLandByZipCode_NotFound(t *testing.T) {
	handler, router := setupTestGinHandler(t)
	seedGinTestData(t, handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/postal-codes/land/zipcode/99999", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp["success"].(bool))
	assert.Contains(t, resp["error"].(string), "not found")
}

func TestGinHandler_GetLandByZipPrefix_Success(t *testing.T) {
	handler, router := setupTestGinHandler(t)
	seedGinTestData(t, handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/postal-codes/land/prefix/256?page=1&limit=10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	assert.NotNil(t, resp["data"])
	assert.Equal(t, float64(2), resp["total"].(float64))
}

func TestGinHandler_SearchLand_Success(t *testing.T) {
	handler, router := setupTestGinHandler(t)
	seedGinTestData(t, handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/postal-codes/land/search?sido_name=강원&page=1&limit=10", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	assert.NotNil(t, resp["data"])
	assert.Equal(t, float64(2), resp["total"].(float64))
}

func TestGinHandler_SearchLand_MultipleParams(t *testing.T) {
	handler, router := setupTestGinHandler(t)
	seedGinTestData(t, handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/postal-codes/land/search?sido_name=강원&eupmyeondong_name=강동면", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
	assert.Equal(t, float64(2), resp["total"].(float64))
}

// ============================================================
// Route Registration Tests
// ============================================================

func TestGinHandler_RegisterGinRoutes(t *testing.T) {
	handler, router := setupTestGinHandler(t)
	seedGinTestData(t, handler)

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{"road search", "GET", "/api/v1/postal-codes/road/search?sido_name=서울", http.StatusOK},
		{"road zipcode", "GET", "/api/v1/postal-codes/road/zipcode/01000", http.StatusOK},
		{"road prefix", "GET", "/api/v1/postal-codes/road/prefix/010", http.StatusOK},
		{"land search", "GET", "/api/v1/postal-codes/land/search?sido_name=강원", http.StatusOK},
		{"land zipcode", "GET", "/api/v1/postal-codes/land/zipcode/25627", http.StatusOK},
		{"land prefix", "GET", "/api/v1/postal-codes/land/prefix/256", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

// ============================================================
// Response Format Tests
// ============================================================

func TestGinHandler_ResponseFormat_Success(t *testing.T) {
	handler, router := setupTestGinHandler(t)
	seedGinTestData(t, handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/postal-codes/road/zipcode/01000", nil)
	router.ServeHTTP(w, req)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Verify response structure
	assert.True(t, resp["success"].(bool))
	assert.NotNil(t, resp["data"])
	_, hasError := resp["error"]
	assert.False(t, hasError)
	assert.Greater(t, resp["total"].(float64), float64(0))

	// Verify content type
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestGinHandler_ResponseFormat_Error(t *testing.T) {
	handler, router := setupTestGinHandler(t)
	seedGinTestData(t, handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/postal-codes/road/zipcode/123", nil)
	router.ServeHTTP(w, req)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	// Verify error response structure
	assert.False(t, resp["success"].(bool))
	_, hasData := resp["data"]
	assert.False(t, hasData)
	assert.NotEmpty(t, resp["error"])

	// Verify content type
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

// ============================================================
// Edge Cases
// ============================================================

func TestGinHandler_EmptyQueryParams(t *testing.T) {
	handler, router := setupTestGinHandler(t)
	seedGinTestData(t, handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/postal-codes/road/search", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
}

func TestGinHandler_InvalidPaginationParams(t *testing.T) {
	handler, router := setupTestGinHandler(t)
	seedGinTestData(t, handler)

	// Invalid page and limit should be ignored and use defaults
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/postal-codes/road/prefix/010?page=invalid&limit=invalid", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["success"].(bool))
}
