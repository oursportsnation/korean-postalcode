package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	postalcode "github.com/oursportsnation/korean-postalcode"
	"github.com/oursportsnation/korean-postalcode/internal/service"
)

// Handler는 우편번호 REST API 핸들러입니다.
type Handler struct {
	service service.Service
}

// New는 새로운 Handler를 생성합니다.
func New(svc service.Service) *Handler {
	return &Handler{service: svc}
}

// Response는 API 응답 구조체입니다.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Total   int64       `json:"total,omitempty"`
}

// RegisterRoutes는 표준 http.ServeMux에 라우트를 등록합니다.
// 사용 예: handler.RegisterRoutes(mux, "/api/v1/postal-codes")
func (h *Handler) RegisterRoutes(mux *http.ServeMux, prefix string) {
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	// 도로명주소 엔드포인트
	mux.HandleFunc(prefix+"road/search", h.Search)
	mux.HandleFunc(prefix+"road/zipcode/", h.GetByZipCode)
	mux.HandleFunc(prefix+"road/prefix/", h.GetByZipPrefix)

	// 지번주소 엔드포인트
	mux.HandleFunc(prefix+"land/search", h.SearchLand)
	mux.HandleFunc(prefix+"land/zipcode/", h.GetLandByZipCode)
	mux.HandleFunc(prefix+"land/prefix/", h.GetLandByZipPrefix)
}

// Search 복합 조건으로 우편번호 검색
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 쿼리 파라미터 파싱
	params := postalcode.SearchParams{
		ZipCode:     r.URL.Query().Get("zip_code"),
		ZipPrefix:   r.URL.Query().Get("zip_prefix"),
		SidoName:    r.URL.Query().Get("sido_name"),
		SigunguName: r.URL.Query().Get("sigungu_name"),
		RoadName:    r.URL.Query().Get("road_name"),
	}

	if page := r.URL.Query().Get("page"); page != "" {
		if val, err := strconv.Atoi(page); err == nil {
			params.Page = val
		}
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil {
			params.Limit = val
		}
	}

	// 검색 실행
	results, total, err := h.service.Search(params)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccess(w, results, total)
}

// GetByZipCode 우편번호로 주소 조회
func (h *Handler) GetByZipCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// URL에서 우편번호 추출 (마지막 경로 세그먼트)
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	zipCode := parts[len(parts)-1]

	// 조회 실행
	results, err := h.service.GetByZipCode(zipCode)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if len(results) == 0 {
		h.sendError(w, http.StatusNotFound, "postal code not found")
		return
	}

	h.sendSuccess(w, results, int64(len(results)))
}

// GetByZipPrefix 우편번호 앞 3자리로 빠른 검색
func (h *Handler) GetByZipPrefix(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// URL에서 우편번호 prefix 추출 (마지막 경로 세그먼트)
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	zipPrefix := parts[len(parts)-1]

	// 페이징 파라미터 파싱
	page := 1   // 기본값
	limit := 10 // 기본값

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if val, err := strconv.Atoi(pageStr); err == nil {
			page = val
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil {
			limit = val
		}
	}

	// page를 offset으로 변환
	offset := (page - 1) * limit

	// 조회 실행
	results, total, err := h.service.GetByZipPrefix(zipPrefix, limit, offset)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.sendSuccess(w, results, total)
}

// sendSuccess는 성공 응답을 보냅니다.
func (h *Handler) sendSuccess(w http.ResponseWriter, data interface{}, total int64) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Response{
		Success: true,
		Data:    data,
		Total:   total,
	})
}

// sendError는 에러 응답을 보냅니다.
func (h *Handler) sendError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(Response{
		Success: false,
		Error:   message,
	})
}

// ============================================================
// 지번주소 관련 핸들러
// ============================================================

// SearchLand 복합 조건으로 지번주소 우편번호 검색
func (h *Handler) SearchLand(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// 쿼리 파라미터 파싱
	params := postalcode.SearchParamsLand{
		ZipCode:          r.URL.Query().Get("zip_code"),
		ZipPrefix:        r.URL.Query().Get("zip_prefix"),
		SidoName:         r.URL.Query().Get("sido_name"),
		SigunguName:      r.URL.Query().Get("sigungu_name"),
		EupmyeondongName: r.URL.Query().Get("eupmyeondong_name"),
		RiName:           r.URL.Query().Get("ri_name"),
	}

	if page := r.URL.Query().Get("page"); page != "" {
		if val, err := strconv.Atoi(page); err == nil {
			params.Page = val
		}
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil {
			params.Limit = val
		}
	}

	// 검색 실행
	results, total, err := h.service.SearchLand(params)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.sendSuccess(w, results, total)
}

// GetLandByZipCode 우편번호로 지번주소 조회
func (h *Handler) GetLandByZipCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// URL에서 우편번호 추출 (마지막 경로 세그먼트)
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	zipCode := parts[len(parts)-1]

	// 조회 실행
	results, err := h.service.GetLandByZipCode(zipCode)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	if len(results) == 0 {
		h.sendError(w, http.StatusNotFound, "postal code not found")
		return
	}

	h.sendSuccess(w, results, int64(len(results)))
}

// GetLandByZipPrefix 우편번호 앞 3자리로 지번주소 빠른 검색
func (h *Handler) GetLandByZipPrefix(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.sendError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// URL에서 우편번호 prefix 추출 (마지막 경로 세그먼트)
	parts := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/")
	zipPrefix := parts[len(parts)-1]

	// 페이징 파라미터 파싱
	page := 1   // 기본값
	limit := 10 // 기본값

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if val, err := strconv.Atoi(pageStr); err == nil {
			page = val
		}
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil {
			limit = val
		}
	}

	// page를 offset으로 변환
	offset := (page - 1) * limit

	// 조회 실행
	results, total, err := h.service.GetLandByZipPrefix(zipPrefix, limit, offset)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.sendSuccess(w, results, total)
}
