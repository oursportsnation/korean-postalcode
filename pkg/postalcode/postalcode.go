// Package postalcode provides a convenient public API for the Korean PostalCode library.
//
// This package serves as the main entry point for users of the library, providing
// simple factory functions to create repositories, services, importers, and HTTP handlers
// without needing to import internal packages directly.
//
// Example usage:
//
//	import (
//	    postalcode "github.com/oursportsnation/korean-postalcode/pkg/postalcode"
//	    "gorm.io/driver/mysql"
//	    "gorm.io/gorm"
//	)
//
//	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})
//	repo := postalcode.NewRepository(db)
//	service := postalcode.NewService(repo)
//	handler := postalcode.NewGinHandler(service)
package postalcode

import (
	stdhttp "net/http"

	"github.com/gin-gonic/gin"
	"github.com/oursportsnation/korean-postalcode/internal/http"
	"github.com/oursportsnation/korean-postalcode/internal/importer"
	"github.com/oursportsnation/korean-postalcode/internal/repository"
	"github.com/oursportsnation/korean-postalcode/internal/service"
	"gorm.io/gorm"
)

// ============================================================
// 공개 인터페이스 (Public Interfaces)
// ============================================================

// Repository는 우편번호 데이터 접근 인터페이스입니다.
type Repository = repository.Repository

// Service는 우편번호 비즈니스 로직을 제공합니다.
type Service = service.Service

// Importer는 파일에서 우편번호 데이터를 가져오는 기능을 제공합니다.
type Importer = importer.Importer

// ============================================================
// 공개 팩토리 함수 (Public Factory Functions)
// ============================================================

// NewRepository는 새로운 Repository를 생성합니다.
//
// 사용 예:
//
//	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})
//	repo := postalcode.NewRepository(db)
func NewRepository(db *gorm.DB) Repository {
	return repository.New(db)
}

// NewService는 새로운 Service를 생성합니다.
//
// 사용 예:
//
//	repo := postalcode.NewRepository(db)
//	service := postalcode.NewService(repo)
func NewService(repo Repository) Service {
	return service.New(repo)
}

// NewImporter는 새로운 Importer를 생성합니다.
//
// 사용 예:
//
//	service := postalcode.NewService(repo)
//	importer := postalcode.NewImporter(service)
func NewImporter(svc Service) Importer {
	return importer.New(svc)
}

// RegisterHTTPRoutes는 표준 HTTP 핸들러 라우트를 등록합니다.
//
// 사용 예:
//
//	service := postalcode.NewService(repo)
//	mux := http.NewServeMux()
//	postalcode.RegisterHTTPRoutes(service, mux, "/api/v1/postal-codes")
func RegisterHTTPRoutes(svc Service, mux *stdhttp.ServeMux, prefix string) {
	handler := http.New(svc)
	handler.RegisterRoutes(mux, prefix)
}

// RegisterGinRoutes는 Gin 프레임워크용 라우트를 등록합니다.
//
// 사용 예:
//
//	service := postalcode.NewService(repo)
//	router := gin.Default()
//	postalcode.RegisterGinRoutes(service, router.Group("/api/v1/postal-codes"))
func RegisterGinRoutes(svc Service, rg *gin.RouterGroup) {
	handler := http.NewGin(svc)
	handler.RegisterGinRoutes(rg)
}
