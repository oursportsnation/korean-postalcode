package http

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	postalcode "github.com/oursportsnation/korean-postalcode"
	"github.com/oursportsnation/korean-postalcode/internal/service"
)

// ErrorResponse는 에러 응답 구조체입니다.
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"잘못된 요청"`
}

// SearchResponse는 검색 응답 구조체입니다.
type SearchResponse struct {
	Success bool                        `json:"success" example:"true"`
	Data    []postalcode.PostalCodeRoad `json:"data"`
	Total   int64                       `json:"total" example:"10"`
}

// SearchResponseLand는 지번주소 검색 응답 구조체입니다.
type SearchResponseLand struct {
	Success bool                        `json:"success" example:"true"`
	Data    []postalcode.PostalCodeLand `json:"data"`
	Total   int64                       `json:"total" example:"10"`
}

// GinHandler는 Gin 프레임워크용 우편번호 API 핸들러입니다.
type GinHandler struct {
	service service.Service
}

// NewGin는 새로운 GinHandler를 생성합니다.
func NewGin(svc service.Service) *GinHandler {
	return &GinHandler{service: svc}
}

// RegisterGinRoutes는 Gin RouterGroup에 라우트를 등록합니다.
// 사용 예: handler.RegisterGinRoutes(router.Group("/api/v1/postal-codes"))
func (h *GinHandler) RegisterGinRoutes(rg *gin.RouterGroup) {
	// 도로명주소 엔드포인트
	road := rg.Group("/road")
	{
		road.GET("/search", h.Search)
		road.GET("/zipcode/:code", h.GetByZipCode)
		road.GET("/prefix/:prefix", h.GetByZipPrefix)
	}

	// 지번주소 엔드포인트
	land := rg.Group("/land")
	{
		land.GET("/search", h.SearchLand)
		land.GET("/zipcode/:code", h.GetLandByZipCode)
		land.GET("/prefix/:prefix", h.GetLandByZipPrefix)
	}
}

// Search godoc
// @Summary 복합 조건으로 우편번호 검색
// @Description 시도, 시군구, 도로명, 우편번호 등 여러 조건으로 검색 가능
// @Tags PostalCodeRoad
// @Accept json
// @Produce json
// @Param zip_code query string false "우편번호 (5자리 정확 매칭)"
// @Param zip_prefix query string false "우편번호 앞 3자리 (권장, 빠른 검색)"
// @Param sido_name query string false "시도명 (부분 매칭)" example("서울특별시")
// @Param sigungu_name query string false "시군구명 (부분 매칭)" example("강북구")
// @Param road_name query string false "도로명 (부분 매칭)" example("삼양로")
// @Param page query int false "페이지 번호 (기본 1)" default(1)
// @Param limit query int false "페이지당 결과 개수 (기본 10, 최대 100)" default(10)
// @Success 200 {object} SearchResponse "성공"
// @Failure 400 {object} ErrorResponse "잘못된 요청"
// @Failure 500 {object} ErrorResponse "서버 오류"
// @Router /api/v1/postal-codes/road/search [get]
func (h *GinHandler) Search(c *gin.Context) {
	params := postalcode.SearchParams{
		ZipCode:     c.Query("zip_code"),
		ZipPrefix:   c.Query("zip_prefix"),
		SidoName:    c.Query("sido_name"),
		SigunguName: c.Query("sigungu_name"),
		RoadName:    c.Query("road_name"),
	}

	if page := c.Query("page"); page != "" {
		if val, err := strconv.Atoi(page); err == nil {
			params.Page = val
		}
	}
	if limit := c.Query("limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil {
			params.Limit = val
		}
	}

	results, total, err := h.service.Search(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
		"total":   total,
	})
}

// GetByZipCode godoc
// @Summary 우편번호로 주소 조회
// @Description 5자리 우편번호로 정확히 매칭되는 도로명 주소 조회
// @Tags PostalCodeRoad
// @Accept json
// @Produce json
// @Param code path string true "우편번호 (5자리)" example("01000")
// @Success 200 {object} SearchResponse "성공"
// @Failure 400 {object} ErrorResponse "잘못된 요청"
// @Failure 404 {object} ErrorResponse "우편번호를 찾을 수 없음"
// @Router /api/v1/postal-codes/road/zipcode/{code} [get]
func (h *GinHandler) GetByZipCode(c *gin.Context) {
	code := c.Param("code")
	results, err := h.service.GetByZipCode(code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if len(results) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "postal code not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
		"total":   int64(len(results)),
	})
}

// GetByZipPrefix godoc
// @Summary 우편번호 앞 3자리로 빠른 검색 (권장)
// @Description 우편번호 앞 3자리로 검색 (인덱스 최적화로 3-5배 빠름)
// @Tags PostalCodeRoad
// @Accept json
// @Produce json
// @Param prefix path string true "우편번호 앞 3자리" example("010")
// @Param page query int false "페이지 번호 (기본 1)" default(1)
// @Param limit query int false "페이지당 결과 개수 (기본 10, 최대 100)" default(10)
// @Success 200 {object} SearchResponse "성공"
// @Failure 400 {object} ErrorResponse "잘못된 요청"
// @Router /api/v1/postal-codes/road/prefix/{prefix} [get]
func (h *GinHandler) GetByZipPrefix(c *gin.Context) {
	prefix := c.Param("prefix")

	// 페이징 파라미터 파싱
	page := 1   // 기본값
	limit := 10 // 기본값

	if pageStr := c.Query("page"); pageStr != "" {
		if val, err := strconv.Atoi(pageStr); err == nil {
			page = val
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil {
			limit = val
		}
	}

	// page를 offset으로 변환
	offset := (page - 1) * limit

	results, total, err := h.service.GetByZipPrefix(prefix, limit, offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
		"total":   total,
	})
}

// ============================================================
// 지번주소 관련 핸들러
// ============================================================

// SearchLand godoc
// @Summary 복합 조건으로 지번주소 우편번호 검색
// @Description 시도, 시군구, 읍면동, 리명, 우편번호 등 여러 조건으로 검색 가능
// @Tags PostalCodeLand
// @Accept json
// @Produce json
// @Param zip_code query string false "우편번호 (5자리 정확 매칭)"
// @Param zip_prefix query string false "우편번호 앞 3자리 (권장, 빠른 검색)"
// @Param sido_name query string false "시도명 (부분 매칭)" example("강원특별자치도")
// @Param sigungu_name query string false "시군구명 (부분 매칭)" example("강릉시")
// @Param eupmyeondong_name query string false "읍면동명 (부분 매칭)" example("강동면")
// @Param ri_name query string false "리명 (부분 매칭)" example("모전리")
// @Param page query int false "페이지 번호 (기본 1)" default(1)
// @Param limit query int false "페이지당 결과 개수 (기본 10, 최대 100)" default(10)
// @Success 200 {object} SearchResponseLand "성공"
// @Failure 400 {object} ErrorResponse "잘못된 요청"
// @Failure 500 {object} ErrorResponse "서버 오류"
// @Router /api/v1/postal-codes/land/search [get]
func (h *GinHandler) SearchLand(c *gin.Context) {
	params := postalcode.SearchParamsLand{
		ZipCode:          c.Query("zip_code"),
		ZipPrefix:        c.Query("zip_prefix"),
		SidoName:         c.Query("sido_name"),
		SigunguName:      c.Query("sigungu_name"),
		EupmyeondongName: c.Query("eupmyeondong_name"),
		RiName:           c.Query("ri_name"),
	}

	if page := c.Query("page"); page != "" {
		if val, err := strconv.Atoi(page); err == nil {
			params.Page = val
		}
	}
	if limit := c.Query("limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil {
			params.Limit = val
		}
	}

	results, total, err := h.service.SearchLand(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
		"total":   total,
	})
}

// GetLandByZipCode godoc
// @Summary 우편번호로 지번주소 조회
// @Description 5자리 우편번호로 정확히 매칭되는 지번주소 조회
// @Tags PostalCodeLand
// @Accept json
// @Produce json
// @Param code path string true "우편번호 (5자리)" example("25627")
// @Success 200 {object} SearchResponseLand "성공"
// @Failure 400 {object} ErrorResponse "잘못된 요청"
// @Failure 404 {object} ErrorResponse "우편번호를 찾을 수 없음"
// @Router /api/v1/postal-codes/land/zipcode/{code} [get]
func (h *GinHandler) GetLandByZipCode(c *gin.Context) {
	code := c.Param("code")
	results, err := h.service.GetLandByZipCode(code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if len(results) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "postal code not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
		"total":   int64(len(results)),
	})
}

// GetLandByZipPrefix godoc
// @Summary 우편번호 앞 3자리로 지번주소 빠른 검색 (권장)
// @Description 우편번호 앞 3자리로 지번주소 검색 (인덱스 최적화로 3-5배 빠름)
// @Tags PostalCodeLand
// @Accept json
// @Produce json
// @Param prefix path string true "우편번호 앞 3자리" example("256")
// @Param page query int false "페이지 번호 (기본 1)" default(1)
// @Param limit query int false "페이지당 결과 개수 (기본 10, 최대 100)" default(10)
// @Success 200 {object} SearchResponseLand "성공"
// @Failure 400 {object} ErrorResponse "잘못된 요청"
// @Router /api/v1/postal-codes/land/prefix/{prefix} [get]
func (h *GinHandler) GetLandByZipPrefix(c *gin.Context) {
	prefix := c.Param("prefix")

	// 페이징 파라미터 파싱
	page := 1   // 기본값
	limit := 10 // 기본값

	if pageStr := c.Query("page"); pageStr != "" {
		if val, err := strconv.Atoi(pageStr); err == nil {
			page = val
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil {
			limit = val
		}
	}

	// page를 offset으로 변환
	offset := (page - 1) * limit

	results, total, err := h.service.GetLandByZipPrefix(prefix, limit, offset)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
		"total":   total,
	})
}
