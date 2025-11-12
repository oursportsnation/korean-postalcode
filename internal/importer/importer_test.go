package importer

import (
	"os"
	"path/filepath"
	"testing"

	postalcode "github.com/oursportsnation/korean-postalcode"
	"github.com/oursportsnation/korean-postalcode/internal/repository"
	"github.com/oursportsnation/korean-postalcode/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestImporter(t *testing.T) Importer {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&postalcode.PostalCodeRoad{}, &postalcode.PostalCodeLand{})
	require.NoError(t, err)

	repo := repository.New(db)
	svc := service.New(repo)
	return New(svc)
}

// ============================================================
// Road Address Import Tests
// ============================================================

func TestImporter_ImportFromFile_Success(t *testing.T) {
	imp := setupTestImporter(t)

	// Get path to test data
	testDataPath := filepath.Join("..", "..", "tests", "testdata", "sample_road.txt")

	// Progress tracking
	var progressCalls int
	progressFn := func(current, total int) {
		progressCalls++
	}

	// Test
	result, err := imp.ImportFromFile(testDataPath, 100, progressFn)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 2, result.TotalCount) // sample_road.txt has 2 data rows
	assert.Equal(t, 0, result.ErrorCount)
	assert.NotEmpty(t, result.Duration)
	assert.Greater(t, progressCalls, 0)
}

func TestImporter_ImportFromFile_FileNotFound(t *testing.T) {
	imp := setupTestImporter(t)

	result, err := imp.ImportFromFile("nonexistent.txt", 100, nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestImporter_ImportFromFile_EmptyFile(t *testing.T) {
	imp := setupTestImporter(t)

	// Create empty temp file
	tmpFile, err := os.CreateTemp("", "empty_*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	result, err := imp.ImportFromFile(tmpFile.Name(), 100, nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestImporter_ImportFromFile_MalformedData(t *testing.T) {
	imp := setupTestImporter(t)

	// Create temp file with malformed data
	tmpFile, err := os.CreateTemp("", "malformed_*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write header + malformed row (missing columns)
	content := `우편번호|시도명|시도명(영문)|시군구명|시군구명(영문)|읍면명|읍면명(영문)|도로명|도로명(영문)|지하여부|건물번호본번(시작)|건물번호부번(시작)|건물번호본번(종료)|건물번호부번(종료)|범위종류
01000|서울특별시|Seoul|강북구
`
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	result, err := imp.ImportFromFile(tmpFile.Name(), 100, nil)
	assert.NoError(t, err) // Import continues despite errors
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.ErrorCount)
}

func TestImporter_ImportFromFile_BatchProcessing(t *testing.T) {
	imp := setupTestImporter(t)

	// Create temp file with multiple rows
	tmpFile, err := os.CreateTemp("", "batch_*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write header + 5 valid rows
	content := `우편번호|시도명|시도명(영문)|시군구명|시군구명(영문)|읍면명|읍면명(영문)|도로명|도로명(영문)|지하여부|건물번호본번(시작)|건물번호부번(시작)|건물번호본번(종료)|건물번호부번(종료)|범위종류
01000|서울특별시|Seoul|강북구|Gangbuk-gu||||||93|0|126|0|3
01001|서울특별시|Seoul|강북구|Gangbuk-gu|||삼양로1|Samyang-ro1|0|1|0|999|0|1
01002|서울특별시|Seoul|강북구|Gangbuk-gu|||삼양로2|Samyang-ro2|0|1|0|999|0|1
01003|서울특별시|Seoul|강북구|Gangbuk-gu|||삼양로3|Samyang-ro3|0|1|0|999|0|1
01004|서울특별시|Seoul|강북구|Gangbuk-gu|||삼양로4|Samyang-ro4|0|1|0|999|0|1
`
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	// Test with batch size of 2
	result, err := imp.ImportFromFile(tmpFile.Name(), 2, nil)
	assert.NoError(t, err)
	assert.Equal(t, 5, result.TotalCount)
	assert.Equal(t, 0, result.ErrorCount)
}

func TestImporter_ImportFromFile_ProgressCallback(t *testing.T) {
	imp := setupTestImporter(t)

	testDataPath := filepath.Join("..", "..", "tests", "testdata", "sample_road.txt")

	var lastCurrent, lastTotal int
	progressFn := func(current, total int) {
		lastCurrent = current
		lastTotal = total
	}

	result, err := imp.ImportFromFile(testDataPath, 100, progressFn)
	assert.NoError(t, err)
	assert.Equal(t, result.TotalCount, lastCurrent)
	assert.Equal(t, result.TotalCount, lastTotal)
}

func TestImporter_ImportFromFile_NilProgressCallback(t *testing.T) {
	imp := setupTestImporter(t)

	testDataPath := filepath.Join("..", "..", "tests", "testdata", "sample_road.txt")

	// Should not panic with nil callback
	result, err := imp.ImportFromFile(testDataPath, 100, nil)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// ============================================================
// Land Address Import Tests
// ============================================================

func TestImporter_ImportLandFromFile_Success(t *testing.T) {
	imp := setupTestImporter(t)

	testDataPath := filepath.Join("..", "..", "tests", "testdata", "sample_land.txt")

	var progressCalls int
	progressFn := func(current, total int) {
		progressCalls++
	}

	result, err := imp.ImportLandFromFile(testDataPath, 100, progressFn)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 3, result.TotalCount) // sample_land.txt has 3 data rows
	assert.Equal(t, 0, result.ErrorCount)
	assert.NotEmpty(t, result.Duration)
	assert.Greater(t, progressCalls, 0)
}

func TestImporter_ImportLandFromFile_FileNotFound(t *testing.T) {
	imp := setupTestImporter(t)

	result, err := imp.ImportLandFromFile("nonexistent.txt", 100, nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestImporter_ImportLandFromFile_MalformedData(t *testing.T) {
	imp := setupTestImporter(t)

	// Create temp file with malformed data
	tmpFile, err := os.CreateTemp("", "malformed_land_*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write header + malformed row (missing columns)
	content := `우편번호|시도명|시도명(영문)|시군구명|시군구명(영문)|읍면동명|읍면동명(영문)|리명|산여부|행정동명|지번본번(시작)|지번부번(시작)|지번본번(종료)|지번부번(종료)
25627|강원특별자치도|Gangwon-do|강릉시
`
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	result, err := imp.ImportLandFromFile(tmpFile.Name(), 100, nil)
	assert.NoError(t, err) // Import continues despite errors
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.ErrorCount)
}

func TestImporter_ImportLandFromFile_BatchProcessing(t *testing.T) {
	imp := setupTestImporter(t)

	// Create temp file with multiple rows
	tmpFile, err := os.CreateTemp("", "batch_land_*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Write header + 5 valid rows
	content := `우편번호|시도명|시도명(영문)|시군구명|시군구명(영문)|읍면동명|읍면동명(영문)|리명|산여부|행정동명|지번본번(시작)|지번부번(시작)|지번본번(종료)|지번부번(종료)
25627|강원특별자치도|Gangwon-do|강릉시|Gangneung-si|강동면|Gangdong-myeon|모전리1|0||2|3|878|0
25628|강원특별자치도|Gangwon-do|강릉시|Gangneung-si|강동면|Gangdong-myeon|모전리2|0||2|3|878|0
25629|강원특별자치도|Gangwon-do|강릉시|Gangneung-si|강동면|Gangdong-myeon|모전리3|0||2|3|878|0
25630|강원특별자치도|Gangwon-do|강릉시|Gangneung-si|강동면|Gangdong-myeon|모전리4|0||2|3|878|0
25631|강원특별자치도|Gangwon-do|강릉시|Gangneung-si|강동면|Gangdong-myeon|모전리5|0||2|3|878|0
`
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	// Test with batch size of 2
	result, err := imp.ImportLandFromFile(tmpFile.Name(), 2, nil)
	assert.NoError(t, err)
	assert.Equal(t, 5, result.TotalCount)
	assert.Equal(t, 0, result.ErrorCount)
}

func TestImporter_ImportLandFromFile_ProgressCallback(t *testing.T) {
	imp := setupTestImporter(t)

	testDataPath := filepath.Join("..", "..", "tests", "testdata", "sample_land.txt")

	var lastCurrent, lastTotal int
	progressFn := func(current, total int) {
		lastCurrent = current
		lastTotal = total
	}

	result, err := imp.ImportLandFromFile(testDataPath, 100, progressFn)
	assert.NoError(t, err)
	assert.Equal(t, result.TotalCount, lastCurrent)
	assert.Equal(t, result.TotalCount, lastTotal)
}

// ============================================================
// Edge Cases and Error Handling
// ============================================================

func TestImporter_InvalidBatchSize(t *testing.T) {
	imp := setupTestImporter(t)

	testDataPath := filepath.Join("..", "..", "tests", "testdata", "sample_road.txt")

	// Test with batch size of 0 or negative
	result, err := imp.ImportFromFile(testDataPath, 0, nil)
	assert.NoError(t, err) // Should still work with default batch size
	assert.NotNil(t, result)

	result, err = imp.ImportFromFile(testDataPath, -10, nil)
	assert.NoError(t, err) // Should still work with default batch size
	assert.NotNil(t, result)
}

func TestImporter_SpecialCharactersInData(t *testing.T) {
	imp := setupTestImporter(t)

	// Create temp file with special characters
	tmpFile, err := os.CreateTemp("", "special_*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := `우편번호|시도명|시도명(영문)|시군구명|시군구명(영문)|읍면명|읍면명(영문)|도로명|도로명(영문)|지하여부|건물번호본번(시작)|건물번호부번(시작)|건물번호본번(종료)|건물번호부번(종료)|범위종류
01000|서울특별시 (Seoul)|Seoul & Korea|강북구!@#|Gangbuk-gu|||특수문자로!|Special-ro|0|93|0|126|0|3
`
	_, err = tmpFile.WriteString(content)
	require.NoError(t, err)
	tmpFile.Close()

	result, err := imp.ImportFromFile(tmpFile.Name(), 100, nil)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.TotalCount)
}
