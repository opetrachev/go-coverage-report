package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	cov, err := ParseCoverage("testdata/01-new-coverage.txt", nil)
	require.NoError(t, err)

	assert.EqualValues(t, 102, cov.TotalStmt)
	assert.EqualValues(t, 92, cov.CoveredStmt)
	assert.EqualValues(t, 10, cov.MissedStmt)
	assert.InDelta(t, 90.196, cov.Percent(), 0.001)
}

func TestCoverage_ByPackage(t *testing.T) {
	cov, err := ParseCoverage("testdata/01-new-coverage.txt", nil)
	require.NoError(t, err)

	pkgs := cov.ByPackage()
	assert.Len(t, pkgs, 1)

	pkgCov := pkgs["github.com/fgrosse/prioqueue"]
	assert.NotNil(t, pkgCov)
	assert.EqualValues(t, 102, pkgCov.TotalStmt)
	assert.EqualValues(t, 92, pkgCov.CoveredStmt)
	assert.EqualValues(t, 10, pkgCov.MissedStmt)
}

func TestParseCoverage_WithExclusions(t *testing.T) {
	// Test excluding specific files
	excludePatterns := []string{"*_test.go", "github.com/fgrosse/prioqueue/min_heap.go"}

	cov, err := ParseCoverage("testdata/01-new-coverage.txt", excludePatterns)
	require.NoError(t, err)

	// Should have fewer files than without exclusions
	assert.Less(t, len(cov.Files), 5) // Original has more files

	// min_heap.go should be excluded
	_, exists := cov.Files["github.com/fgrosse/prioqueue/min_heap.go"]
	assert.False(t, exists, "min_heap.go should be excluded")

	// Test files should be excluded
	for filename := range cov.Files {
		assert.False(t, strings.HasSuffix(filename, "_test.go"), "Test files should be excluded")
	}
}

func TestParseChangedFiles_WithExclusions(t *testing.T) {
	// Test excluding specific files from changed files
	excludePatterns := []string{"github.com/fgrosse/prioqueue/min_heap.go"}

	changedFiles, err := ParseChangedFiles("testdata/01-changed-files.json", "github.com/fgrosse/prioqueue", excludePatterns)
	require.NoError(t, err)

	// Should have only one file (foo/bar/baz.go), min_heap.go should be excluded
	assert.Len(t, changedFiles, 1)
	assert.Equal(t, "github.com/fgrosse/prioqueue/foo/bar/baz.go", changedFiles[0])

	// min_heap.go should not be in the list
	for _, file := range changedFiles {
		assert.NotEqual(t, "github.com/fgrosse/prioqueue/min_heap.go", file, "min_heap.go should be excluded from changed files")
	}
}
