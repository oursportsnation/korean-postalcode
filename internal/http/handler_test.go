package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	postalcode "github.com/oursportsnation/korean-postalcode"
	"github.com/oursportsnation/korean-postalcode/internal/repository"
	"github.com/oursportsnation/korean-postalcode/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestHandler(t *testing.T) *Handler {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&postalcode.PostalCodeRoad{}, &postalcode.PostalCodeLand{})
	require.NoError(t, err)

	repo := repository.New(db)
	svc := service.New(repo)
	return New(svc)
}

func seedTestData(t *testing.T, handler *Handler) {
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
// Road Address Handler Tests
// ============================================================

func TestHandler_GetByZipCode_Success(t *testing.T) {
	handler := setupTestHandler(t)
	seedTestData(t, handler)

	req := httptest.NewRequest("GET", "/road/zipcode/01000", nil)
	w := httptest.NewRecorder()

	handler.GetByZipCode(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Data)
	assert.Equal(t, int64(1), resp.Total)
}

func TestHandler_GetByZipCode_NotFound(t *testing.T) {
	handler := setupTestHandler(t)
	seedTestData(t, handler)

	req := httptest.NewRequest("GET", "/road/zipcode/99999", nil)
	w := httptest.NewRecorder()

	handler.GetByZipCode(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "not found")
}

func TestHandler_GetByZipCode_InvalidZipCode(t *testing.T) {
	handler := setupTestHandler(t)
	seedTestData(t, handler)

	req := httptest.NewRequest("GET", "/road/zipcode/123", nil)
	w := httptest.NewRecorder()

	handler.GetByZipCode(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.NotEmpty(t, resp.Error)
}

func TestHandler_GetByZipCode_MethodNotAllowed(t *testing.T) {
	handler := setupTestHandler(t)

	req := httptest.NewRequest("POST", "/road/zipcode/01000", nil)
	w := httptest.NewRecorder()

	handler.GetByZipCode(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	var resp Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "method not allowed")
}

func TestHandler_GetByZipPrefix_Success(t *testing.T) {
	handler := setupTestHandler(t)
	seedTestData(t, handler)

	req := httptest.NewRequest("GET", "/road/prefix/010?page=1&limit=10", nil)
	w := httptest.NewRecorder()

	handler.GetByZipPrefix(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Data)
	assert.Equal(t, int64(2), resp.Total)
}

func TestHandler_GetByZipPrefix_Pagination(t *testing.T) {
	handler := setupTestHandler(t)
	seedTestData(t, handler)

	// Test page 1 with limit 1
	req := httptest.NewRequest("GET", "/road/prefix/010?page=1&limit=1", nil)
	w := httptest.NewRecorder()

	handler.GetByZipPrefix(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, int64(2), resp.Total) // Total count should be 2

	// Verify data is a slice with 1 item
	data, ok := resp.Data.([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 1)
}

func TestHandler_GetByZipPrefix_InvalidPrefix(t *testing.T) {
	handler := setupTestHandler(t)
	seedTestData(t, handler)

	req := httptest.NewRequest("GET", "/road/prefix/12", nil)
	w := httptest.NewRecorder()

	handler.GetByZipPrefix(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.NotEmpty(t, resp.Error)
}

func TestHandler_Search_Success(t *testing.T) {
	handler := setupTestHandler(t)
	seedTestData(t, handler)

	req := httptest.NewRequest("GET", "/road/search?sido_name=서울&page=1&limit=10", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Data)
	assert.Equal(t, int64(3), resp.Total) // All 3 Seoul addresses
}

func TestHandler_Search_MultipleParams(t *testing.T) {
	handler := setupTestHandler(t)
	seedTestData(t, handler)

	req := httptest.NewRequest("GET", "/road/search?sido_name=서울&sigungu_name=강북", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, int64(2), resp.Total) // 2 Gangbuk addresses
}

func TestHandler_Search_NoResults(t *testing.T) {
	handler := setupTestHandler(t)
	seedTestData(t, handler)

	req := httptest.NewRequest("GET", "/road/search?sido_name=부산", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, int64(0), resp.Total)
}

func TestHandler_Search_MethodNotAllowed(t *testing.T) {
	handler := setupTestHandler(t)

	req := httptest.NewRequest("POST", "/road/search", nil)
	w := httptest.NewRecorder()

	handler.Search(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	var resp Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
}

// ============================================================
// Land Address Handler Tests
// ============================================================

func TestHandler_GetLandByZipCode_Success(t *testing.T) {
	handler := setupTestHandler(t)
	seedTestData(t, handler)

	req := httptest.NewRequest("GET", "/land/zipcode/25627", nil)
	w := httptest.NewRecorder()

	handler.GetLandByZipCode(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Data)
	assert.Equal(t, int64(1), resp.Total)
}

func TestHandler_GetLandByZipCode_NotFound(t *testing.T) {
	handler := setupTestHandler(t)
	seedTestData(t, handler)

	req := httptest.NewRequest("GET", "/land/zipcode/99999", nil)
	w := httptest.NewRecorder()

	handler.GetLandByZipCode(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "not found")
}

func TestHandler_GetLandByZipPrefix_Success(t *testing.T) {
	handler := setupTestHandler(t)
	seedTestData(t, handler)

	req := httptest.NewRequest("GET", "/land/prefix/256?page=1&limit=10", nil)
	w := httptest.NewRecorder()

	handler.GetLandByZipPrefix(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Data)
	assert.Equal(t, int64(2), resp.Total)
}

func TestHandler_SearchLand_Success(t *testing.T) {
	handler := setupTestHandler(t)
	seedTestData(t, handler)

	req := httptest.NewRequest("GET", "/land/search?sido_name=강원&page=1&limit=10", nil)
	w := httptest.NewRecorder()

	handler.SearchLand(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Data)
	assert.Equal(t, int64(2), resp.Total)
}

func TestHandler_SearchLand_MultipleParams(t *testing.T) {
	handler := setupTestHandler(t)
	seedTestData(t, handler)

	req := httptest.NewRequest("GET", "/land/search?sido_name=강원&eupmyeondong_name=강동면", nil)
	w := httptest.NewRecorder()

	handler.SearchLand(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.True(t, resp.Success)
	assert.Equal(t, int64(2), resp.Total)
}

func TestHandler_SearchLand_MethodNotAllowed(t *testing.T) {
	handler := setupTestHandler(t)

	req := httptest.NewRequest("POST", "/land/search", nil)
	w := httptest.NewRecorder()

	handler.SearchLand(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	var resp Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	assert.False(t, resp.Success)
}

// ============================================================
// Route Registration Tests
// ============================================================

func TestHandler_RegisterRoutes(t *testing.T) {
	handler := setupTestHandler(t)
	seedTestData(t, handler)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, "/api/v1/postal-codes")

	// Test road endpoints
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
			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestHandler_RegisterRoutes_WithoutTrailingSlash(t *testing.T) {
	handler := setupTestHandler(t)
	seedTestData(t, handler)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux, "/api/v1/postal-codes") // No trailing slash

	req := httptest.NewRequest("GET", "/api/v1/postal-codes/road/zipcode/01000", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// ============================================================
// Response Format Tests
// ============================================================

func TestHandler_ResponseFormat_Success(t *testing.T) {
	handler := setupTestHandler(t)
	seedTestData(t, handler)

	req := httptest.NewRequest("GET", "/road/zipcode/01000", nil)
	w := httptest.NewRecorder()

	handler.GetByZipCode(w, req)

	var resp Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	// Verify response structure
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Data)
	assert.Empty(t, resp.Error)
	assert.Greater(t, resp.Total, int64(0))

	// Verify content type
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}

func TestHandler_ResponseFormat_Error(t *testing.T) {
	handler := setupTestHandler(t)

	req := httptest.NewRequest("GET", "/road/zipcode/123", nil)
	w := httptest.NewRecorder()

	handler.GetByZipCode(w, req)

	var resp Response
	err := json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)

	// Verify error response structure
	assert.False(t, resp.Success)
	assert.Nil(t, resp.Data)
	assert.NotEmpty(t, resp.Error)

	// Verify content type
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
}
