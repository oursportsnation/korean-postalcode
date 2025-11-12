package repository

import (
	"testing"

	postalcode "github.com/oursportsnation/korean-postalcode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto migrate
	err = db.AutoMigrate(&postalcode.PostalCodeRoad{}, &postalcode.PostalCodeLand{})
	require.NoError(t, err)

	return db
}

func TestRepository_Road_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)

	road := &postalcode.PostalCodeRoad{
		ZipCode:           "01000",
		ZipPrefix:         "010",
		SidoName:          "서울특별시",
		SidoNameEn:        "Seoul",
		SigunguName:       "강북구",
		SigunguNameEn:     "Gangbuk-gu",
		RoadName:          "삼양로177길",
		RoadNameEn:        "Samyang-ro 177-gil",
		IsUnderground:     false,
		StartBuildingMain: 93,
		RangeType:         3,
	}

	err := repo.Create(road)
	assert.NoError(t, err)
	assert.NotZero(t, road.ID)
}

func TestRepository_Road_FindByZipCode(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)

	// Seed data
	road := &postalcode.PostalCodeRoad{
		ZipCode:     "01000",
		ZipPrefix:   "010",
		SidoName:    "서울특별시",
		SigunguName: "강북구",
		RoadName:    "삼양로177길",
	}
	require.NoError(t, repo.Create(road))

	// Test
	results, err := repo.FindByZipCode("01000")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "01000", results[0].ZipCode)
	assert.Equal(t, "서울특별시", results[0].SidoName)
}

func TestRepository_Road_FindByZipPrefix(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)

	// Seed data
	roads := []postalcode.PostalCodeRoad{
		{ZipCode: "01000", ZipPrefix: "010", SidoName: "서울특별시", SigunguName: "강북구", RoadName: "삼양로1"},
		{ZipCode: "01001", ZipPrefix: "010", SidoName: "서울특별시", SigunguName: "강북구", RoadName: "삼양로2"},
		{ZipCode: "06000", ZipPrefix: "060", SidoName: "서울특별시", SigunguName: "강남구", RoadName: "테헤란로"},
	}
	for i := range roads {
		require.NoError(t, repo.Create(&roads[i]))
	}

	// Test
	results, total, err := repo.FindByZipPrefix("010", 10, 0)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, results, 2)
}

func TestRepository_Road_Search(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)

	// Seed data
	roads := []postalcode.PostalCodeRoad{
		{ZipCode: "01000", ZipPrefix: "010", SidoName: "서울특별시", SigunguName: "강북구", RoadName: "삼양로"},
		{ZipCode: "06000", ZipPrefix: "060", SidoName: "서울특별시", SigunguName: "강남구", RoadName: "테헤란로"},
		{ZipCode: "21000", ZipPrefix: "210", SidoName: "부산광역시", SigunguName: "중구", RoadName: "중앙대로"},
	}
	for i := range roads {
		require.NoError(t, repo.Create(&roads[i]))
	}

	// Test: Search by SidoName
	params := postalcode.SearchParams{
		SidoName: "서울",
		Page:     1,
		Limit:    10,
	}
	results, total, err := repo.Search(params)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, results, 2)

	// Test: Search by SigunguName
	params = postalcode.SearchParams{
		SigunguName: "강북",
		Page:        1,
		Limit:       10,
	}
	results, total, err = repo.Search(params)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, results, 1)
	assert.Equal(t, "강북구", results[0].SigunguName)
}

func TestRepository_Road_BatchCreate(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)

	roads := []postalcode.PostalCodeRoad{
		{ZipCode: "01000", ZipPrefix: "010", SidoName: "서울특별시", SigunguName: "강북구", RoadName: "삼양로1"},
		{ZipCode: "01001", ZipPrefix: "010", SidoName: "서울특별시", SigunguName: "강북구", RoadName: "삼양로2"},
		{ZipCode: "06000", ZipPrefix: "060", SidoName: "서울특별시", SigunguName: "강남구", RoadName: "테헤란로"},
	}

	err := repo.BatchCreate(roads)
	assert.NoError(t, err)

	// Verify
	var count int64
	db.Model(&postalcode.PostalCodeRoad{}).Count(&count)
	assert.Equal(t, int64(3), count)
}

func TestRepository_Road_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)

	// Create
	road := &postalcode.PostalCodeRoad{
		ZipCode:     "01000",
		ZipPrefix:   "010",
		SidoName:    "서울특별시",
		SigunguName: "강북구",
		RoadName:    "삼양로177길",
	}
	require.NoError(t, repo.Create(road))

	// Update
	road.RoadName = "Updated Road"
	err := repo.Update(road)
	assert.NoError(t, err)

	// Verify
	results, err := repo.FindByZipCode("01000")
	require.NoError(t, err)
	assert.Equal(t, "Updated Road", results[0].RoadName)
}

func TestRepository_Road_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)

	// Create
	road := &postalcode.PostalCodeRoad{
		ZipCode:     "01000",
		ZipPrefix:   "010",
		SidoName:    "서울특별시",
		SigunguName: "강북구",
		RoadName:    "삼양로177길",
	}
	require.NoError(t, repo.Create(road))

	// Delete
	err := repo.Delete(road.ID)
	assert.NoError(t, err)

	// Verify
	results, err := repo.FindByZipCode("01000")
	assert.NoError(t, err)
	assert.Empty(t, results)
}

// ============================================================
// Land Address Tests
// ============================================================

func TestRepository_Land_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)

	land := &postalcode.PostalCodeLand{
		ZipCode:            "25627",
		ZipPrefix:          "256",
		SidoName:           "강원특별자치도",
		SidoNameEn:         "Gangwon-do",
		SigunguName:        "강릉시",
		SigunguNameEn:      "Gangneung-si",
		EupmyeondongName:   "강동면",
		EupmyeondongNameEn: "Gangdong-myeon",
		RiName:             "모전리",
		IsMountain:         false,
		StartJibunMain:     2,
	}

	err := repo.CreateLand(land)
	assert.NoError(t, err)
	assert.NotZero(t, land.ID)
}

func TestRepository_Land_FindByZipCode(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)

	// Seed data
	land := &postalcode.PostalCodeLand{
		ZipCode:          "25627",
		ZipPrefix:        "256",
		SidoName:         "강원특별자치도",
		SigunguName:      "강릉시",
		EupmyeondongName: "강동면",
		RiName:           "모전리",
	}
	require.NoError(t, repo.CreateLand(land))

	// Test
	results, err := repo.FindLandByZipCode("25627")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "25627", results[0].ZipCode)
	assert.Equal(t, "강원특별자치도", results[0].SidoName)
}

func TestRepository_Land_SearchLand(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)

	// Seed data
	lands := []postalcode.PostalCodeLand{
		{ZipCode: "25627", ZipPrefix: "256", SidoName: "강원특별자치도", SigunguName: "강릉시", EupmyeondongName: "강동면", RiName: "모전리"},
		{ZipCode: "25628", ZipPrefix: "256", SidoName: "강원특별자치도", SigunguName: "강릉시", EupmyeondongName: "강동면", RiName: "심곡리"},
		{ZipCode: "48000", ZipPrefix: "480", SidoName: "부산광역시", SigunguName: "중구", EupmyeondongName: "중앙동", RiName: ""},
	}
	for i := range lands {
		require.NoError(t, repo.CreateLand(&lands[i]))
	}

	// Test: Search by SidoName
	params := postalcode.SearchParamsLand{
		SidoName: "강원",
		Page:     1,
		Limit:    10,
	}
	results, total, err := repo.SearchLand(params)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, results, 2)

	// Test: Search by EupmyeondongName
	params = postalcode.SearchParamsLand{
		EupmyeondongName: "강동면",
		Page:             1,
		Limit:            10,
	}
	results, total, err = repo.SearchLand(params)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, results, 2)
}

func TestRepository_Land_BatchCreate(t *testing.T) {
	db := setupTestDB(t)
	repo := New(db)

	lands := []postalcode.PostalCodeLand{
		{ZipCode: "25627", ZipPrefix: "256", SidoName: "강원특별자치도", SigunguName: "강릉시", EupmyeondongName: "강동면", RiName: "모전리"},
		{ZipCode: "25628", ZipPrefix: "256", SidoName: "강원특별자치도", SigunguName: "강릉시", EupmyeondongName: "강동면", RiName: "심곡리"},
	}

	err := repo.BatchCreateLand(lands)
	assert.NoError(t, err)

	// Verify
	var count int64
	db.Model(&postalcode.PostalCodeLand{}).Count(&count)
	assert.Equal(t, int64(2), count)
}
