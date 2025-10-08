package coveragechanges

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing
type mockCoverage struct {
	files map[string]FileProfile
}

func (mc *mockCoverage) GetFiles() map[string]FileProfile {
	return mc.files
}

type mockFileProfile struct {
	percent     float64
	totalStmt   int64
	coveredStmt int64
	missedStmt  int64
}

func (mfp *mockFileProfile) CoveragePercent() float64 {
	return mfp.percent
}

func (mfp *mockFileProfile) GetTotal() int64 {
	return mfp.totalStmt
}

func (mfp *mockFileProfile) GetCovered() int64 {
	return mfp.coveredStmt
}

func (mfp *mockFileProfile) GetMissed() int64 {
	return mfp.missedStmt
}

func TestNew(t *testing.T) {
	oldCov := &mockCoverage{
		files: map[string]FileProfile{
			"pkg/file1.go": &mockFileProfile{
				percent:     80.0,
				totalStmt:   100,
				coveredStmt: 80,
				missedStmt:  20,
			},
			"pkg/file2.go": &mockFileProfile{
				percent:     60.0,
				totalStmt:   50,
				coveredStmt: 30,
				missedStmt:  20,
			},
		},
	}

	newCov := &mockCoverage{
		files: map[string]FileProfile{
			"pkg/file1.go": &mockFileProfile{
				percent:     85.0,
				totalStmt:   100,
				coveredStmt: 85,
				missedStmt:  15,
			},
			"pkg/file2.go": &mockFileProfile{
				percent:     70.0,
				totalStmt:   50,
				coveredStmt: 35,
				missedStmt:  15,
			},
			"pkg/file3.go": &mockFileProfile{
				percent:     90.0,
				totalStmt:   20,
				coveredStmt: 18,
				missedStmt:  2,
			},
		},
	}

	changes := New(oldCov, newCov)
	require.NotNil(t, changes)

	// Test file1 - exists in both
	file1Change := changes.GetFileChange("pkg/file1.go")
	require.NotNil(t, file1Change)
	assert.Equal(t, "pkg/file1.go", file1Change.FileName)
	assert.Equal(t, 80.0, file1Change.GetOldPercent())
	assert.Equal(t, 85.0, file1Change.GetNewPercent())
	assert.Equal(t, 5.0, file1Change.GetDelta())

	// Test file3 - only in new coverage
	file3Change := changes.GetFileChange("pkg/file3.go")
	require.NotNil(t, file3Change)
	assert.Equal(t, "pkg/file3.go", file3Change.FileName)
	assert.Equal(t, 0.0, file3Change.GetOldPercent())
	assert.Equal(t, 90.0, file3Change.GetNewPercent())
	assert.Equal(t, 90.0, file3Change.GetDelta())

	// Test GetAllFiles
	allFiles := changes.GetAllFiles()
	assert.Len(t, allFiles, 3)
	assert.Contains(t, allFiles, "pkg/file1.go")
	assert.Contains(t, allFiles, "pkg/file2.go")
	assert.Contains(t, allFiles, "pkg/file3.go")
}

func TestGetPackageChange(t *testing.T) {
	oldCov := &mockCoverage{
		files: map[string]FileProfile{
			"pkg/file1.go": &mockFileProfile{
				percent:     80.0,
				totalStmt:   100,
				coveredStmt: 80,
				missedStmt:  20,
			},
			"pkg/file2.go": &mockFileProfile{
				percent:     60.0,
				totalStmt:   50,
				coveredStmt: 30,
				missedStmt:  20,
			},
			"other/file3.go": &mockFileProfile{
				percent:     50.0,
				totalStmt:   20,
				coveredStmt: 10,
				missedStmt:  10,
			},
		},
	}

	newCov := &mockCoverage{
		files: map[string]FileProfile{
			"pkg/file1.go": &mockFileProfile{
				percent:     85.0,
				totalStmt:   100,
				coveredStmt: 85,
				missedStmt:  15,
			},
			"pkg/file2.go": &mockFileProfile{
				percent:     70.0,
				totalStmt:   50,
				coveredStmt: 35,
				missedStmt:  15,
			},
			"other/file3.go": &mockFileProfile{
				percent:     60.0,
				totalStmt:   20,
				coveredStmt: 12,
				missedStmt:  8,
			},
		},
	}

	changes := New(oldCov, newCov)

	// Test package "pkg"
	pkgChange := changes.GetPackageChange("pkg")
	require.NotNil(t, pkgChange)
	assert.Equal(t, "pkg", pkgChange.FileName)

	// Old coverage for pkg: (80 + 30) / (100 + 50) = 110/150 = 73.33%
	assert.InDelta(t, 73.33, pkgChange.GetOldPercent(), 0.01)

	// New coverage for pkg: (85 + 35) / (100 + 50) = 120/150 = 80%
	assert.InDelta(t, 80.0, pkgChange.GetNewPercent(), 0.01)

	// Delta: 80 - 73.33 = 6.67%
	assert.InDelta(t, 6.67, pkgChange.GetDelta(), 0.01)

	// Test GetPackages
	packages := changes.GetPackages()
	assert.Len(t, packages, 2)
	assert.Contains(t, packages, "pkg")
	assert.Contains(t, packages, "other")
}

func TestCoverageChange(t *testing.T) {
	// Test positive delta
	assert.Equal(t, 10.0, CoverageChange(80.0, 90.0))

	// Test negative delta
	assert.Equal(t, 10.0, CoverageChange(90.0, 80.0))

	// Test zero delta
	assert.Equal(t, 0.0, CoverageChange(80.0, 80.0))
}
