package renderer

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing
type mockCoverageData struct {
	fileChanges    map[string]*mockFileChange
	packageChanges map[string]*mockFileChange
	allFiles       []string
	packages       []string
}

func (mcd *mockCoverageData) GetFileChange(file string) FileChange {
	if fc, ok := mcd.fileChanges[file]; ok {
		return fc
	}
	return nil
}

func (mcd *mockCoverageData) GetPackageChange(pkg string) FileChange {
	if fc, ok := mcd.packageChanges[pkg]; ok {
		return fc
	}
	return nil
}

func (mcd *mockCoverageData) GetAllFiles() []string {
	return mcd.allFiles
}

func (mcd *mockCoverageData) GetPackages() []string {
	return mcd.packages
}

type mockFileChange struct {
	oldPercent  float64
	newPercent  float64
	oldCoverage *mockFileCoverage
	newCoverage *mockFileCoverage
}

func (mfc *mockFileChange) GetOldPercent() float64 {
	return mfc.oldPercent
}

func (mfc *mockFileChange) GetNewPercent() float64 {
	return mfc.newPercent
}

func (mfc *mockFileChange) GetDelta() float64 {
	return mfc.newPercent - mfc.oldPercent
}

func (mfc *mockFileChange) GetOldCoverage() FileCoverage {
	if mfc.oldCoverage == nil {
		return &mockFileCoverage{}
	}
	return mfc.oldCoverage
}

func (mfc *mockFileChange) GetNewCoverage() FileCoverage {
	if mfc.newCoverage == nil {
		return &mockFileCoverage{}
	}
	return mfc.newCoverage
}

type mockFileCoverage struct {
	percent     float64
	totalStmt   int64
	coveredStmt int64
	missedStmt  int64
}

func (mfc *mockFileCoverage) GetPercent() float64 {
	return mfc.percent
}

func (mfc *mockFileCoverage) GetTotalStmt() int64 {
	return mfc.totalStmt
}

func (mfc *mockFileCoverage) GetCoveredStmt() int64 {
	return mfc.coveredStmt
}

func (mfc *mockFileCoverage) GetMissedStmt() int64 {
	return mfc.missedStmt
}

func TestRenderer_Markdown_HappyPath(t *testing.T) {
	data := &mockCoverageData{
		fileChanges: map[string]*mockFileChange{
			"pkg/file1.go": {
				oldPercent: 80.0,
				newPercent: 85.0,
				oldCoverage: &mockFileCoverage{
					percent:     80.0,
					totalStmt:   100,
					coveredStmt: 80,
					missedStmt:  20,
				},
				newCoverage: &mockFileCoverage{
					percent:     85.0,
					totalStmt:   100,
					coveredStmt: 85,
					missedStmt:  15,
				},
			},
			"pkg/file2.go": {
				oldPercent: 60.0,
				newPercent: 70.0,
				oldCoverage: &mockFileCoverage{
					percent:     60.0,
					totalStmt:   50,
					coveredStmt: 30,
					missedStmt:  20,
				},
				newCoverage: &mockFileCoverage{
					percent:     70.0,
					totalStmt:   50,
					coveredStmt: 35,
					missedStmt:  15,
				},
			},
		},
		packageChanges: map[string]*mockFileChange{
			"pkg": {
				oldPercent: 73.33,
				newPercent: 80.0,
			},
		},
		allFiles: []string{"pkg/file1.go", "pkg/file2.go"},
		packages: []string{"pkg"},
	}

	rend := New(data, Options{
		ChangedFiles:           []string{"pkg/file1.go", "pkg/file2.go"},
		PackageThreshold:       0,
		PackageFileThreshold:   0,
		FileExclusionThreshold: 0,
		ExcludePatterns:        nil,
		TrimPrefix:             "",
	})

	markdown := rend.Markdown()
	require.NotEmpty(t, markdown)

	// Check that the report contains expected sections
	assert.Contains(t, markdown, "### Merging this branch will **increase** overall coverage")
	assert.Contains(t, markdown, "| Impacted Packages | Coverage Δ | :robot: |")
	assert.Contains(t, markdown, "| pkg | 80.00% (**+6.67%**) | :thumbsup: |")
	assert.Contains(t, markdown, "<details>")
	assert.Contains(t, markdown, "Coverage by file")
	assert.Contains(t, markdown, "| Changed File | Coverage Δ | Total | Covered | Missed | :robot: |")
	assert.Contains(t, markdown, "| pkg/file1.go | 85.00% (**+5.00%**) | 100 | 85 (+5) | 15 (-5) | :thumbsup: |")
	assert.Contains(t, markdown, "| pkg/file2.go | 70.00% (**+10.00%**) | 50 | 35 (+5) | 15 (-5) | :thumbsup: |")
}

func TestRenderer_Markdown_WithThresholds(t *testing.T) {
	data := &mockCoverageData{
		fileChanges: map[string]*mockFileChange{
			"pkg/file1.go": {
				oldPercent: 80.0,
				newPercent: 81.0, // Small change (1%)
				oldCoverage: &mockFileCoverage{
					percent:     80.0,
					totalStmt:   100,
					coveredStmt: 80,
					missedStmt:  20,
				},
				newCoverage: &mockFileCoverage{
					percent:     81.0,
					totalStmt:   100,
					coveredStmt: 81,
					missedStmt:  19,
				},
			},
			"pkg/file2.go": {
				oldPercent: 60.0,
				newPercent: 75.0, // Large change (15%)
				oldCoverage: &mockFileCoverage{
					percent:     60.0,
					totalStmt:   50,
					coveredStmt: 30,
					missedStmt:  20,
				},
				newCoverage: &mockFileCoverage{
					percent:     75.0,
					totalStmt:   50,
					coveredStmt: 37,
					missedStmt:  13,
				},
			},
		},
		packageChanges: map[string]*mockFileChange{
			"pkg": {
				oldPercent: 73.33,
				newPercent: 78.67,
			},
		},
		allFiles: []string{"pkg/file1.go", "pkg/file2.go"},
		packages: []string{"pkg"},
	}

	// Test with file exclusion threshold of 10%
	rend := New(data, Options{
		ChangedFiles:           []string{"pkg/file1.go", "pkg/file2.go"},
		PackageThreshold:       0,
		PackageFileThreshold:   0,
		FileExclusionThreshold: 10.0, // Exclude files with change < 10%
		ExcludePatterns:        nil,
		TrimPrefix:             "",
	})

	markdown := rend.Markdown()
	require.NotEmpty(t, markdown)

	// file1.go should be excluded (change is only 1%)
	assert.NotContains(t, markdown, "pkg/file1.go")

	// file2.go should be included (change is 15%)
	assert.Contains(t, markdown, "pkg/file2.go")
}

func TestRenderer_Markdown_WithExclusionPatterns(t *testing.T) {
	data := &mockCoverageData{
		fileChanges: map[string]*mockFileChange{
			"pkg/file1.go": {
				oldPercent: 80.0,
				newPercent: 85.0,
				oldCoverage: &mockFileCoverage{
					percent:     80.0,
					totalStmt:   100,
					coveredStmt: 80,
					missedStmt:  20,
				},
				newCoverage: &mockFileCoverage{
					percent:     85.0,
					totalStmt:   100,
					coveredStmt: 85,
					missedStmt:  15,
				},
			},
			"pkg/file2.go": {
				oldPercent: 60.0,
				newPercent: 70.0,
				oldCoverage: &mockFileCoverage{
					percent:     60.0,
					totalStmt:   50,
					coveredStmt: 30,
					missedStmt:  20,
				},
				newCoverage: &mockFileCoverage{
					percent:     70.0,
					totalStmt:   50,
					coveredStmt: 35,
					missedStmt:  15,
				},
			},
		},
		packageChanges: map[string]*mockFileChange{
			"pkg": {
				oldPercent: 73.33,
				newPercent: 80.0,
			},
		},
		allFiles: []string{"pkg/file1.go", "pkg/file2.go"},
		packages: []string{"pkg"},
	}

	// Test with exclusion pattern for file2.go
	rend := New(data, Options{
		ChangedFiles:           []string{"pkg/file1.go", "pkg/file2.go"},
		PackageThreshold:       0,
		PackageFileThreshold:   0,
		FileExclusionThreshold: 0,
		ExcludePatterns:        []string{"pkg/file2.go"},
		TrimPrefix:             "",
	})

	markdown := rend.Markdown()
	require.NotEmpty(t, markdown)

	// file2.go should be excluded
	assert.NotContains(t, markdown, "pkg/file2.go")

	// file1.go should be included
	assert.Contains(t, markdown, "pkg/file1.go")
}

func TestRenderer_Markdown_WithTrimPrefix(t *testing.T) {
	data := &mockCoverageData{
		fileChanges: map[string]*mockFileChange{
			"github.com/example/pkg/file1.go": {
				oldPercent: 80.0,
				newPercent: 85.0,
				oldCoverage: &mockFileCoverage{
					percent:     80.0,
					totalStmt:   100,
					coveredStmt: 80,
					missedStmt:  20,
				},
				newCoverage: &mockFileCoverage{
					percent:     85.0,
					totalStmt:   100,
					coveredStmt: 85,
					missedStmt:  15,
				},
			},
		},
		packageChanges: map[string]*mockFileChange{
			"github.com/example/pkg": {
				oldPercent: 80.0,
				newPercent: 85.0,
			},
		},
		allFiles: []string{"github.com/example/pkg/file1.go"},
		packages: []string{"github.com/example/pkg"},
	}

	rend := New(data, Options{
		ChangedFiles:           []string{"github.com/example/pkg/file1.go"},
		PackageThreshold:       0,
		PackageFileThreshold:   0,
		FileExclusionThreshold: 0,
		ExcludePatterns:        nil,
		TrimPrefix:             "github.com/example/",
	})

	markdown := rend.Markdown()
	require.NotEmpty(t, markdown)

	// Check that prefix is trimmed
	assert.Contains(t, markdown, "| pkg |")
	assert.Contains(t, markdown, "| pkg/file1.go |")

	// Check that full path is not present
	assert.NotContains(t, markdown, "github.com/example/pkg |")
}

func TestRenderer_Title(t *testing.T) {
	tests := []struct {
		name             string
		packageChanges   map[string]*mockFileChange
		expectedContains string
	}{
		{
			name: "increase",
			packageChanges: map[string]*mockFileChange{
				"pkg": {oldPercent: 70.0, newPercent: 80.0},
			},
			expectedContains: "**increase**",
		},
		{
			name: "decrease",
			packageChanges: map[string]*mockFileChange{
				"pkg": {oldPercent: 80.0, newPercent: 70.0},
			},
			expectedContains: "**decrease**",
		},
		{
			name: "no change",
			packageChanges: map[string]*mockFileChange{
				"pkg": {oldPercent: 80.0, newPercent: 80.0},
			},
			expectedContains: "**not change**",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := &mockCoverageData{
				packageChanges: tt.packageChanges,
				packages:       []string{"pkg"},
			}

			rend := New(data, Options{
				ChangedFiles: []string{"pkg/file.go"},
			})

			title := rend.Title()
			assert.Contains(t, title, tt.expectedContains)
		})
	}
}

func TestRenderer_JSON(t *testing.T) {
	data := &mockCoverageData{
		fileChanges: map[string]*mockFileChange{
			"pkg/file1.go": {
				oldPercent: 80.0,
				newPercent: 85.0,
			},
		},
		packageChanges: map[string]*mockFileChange{
			"pkg": {
				oldPercent: 80.0,
				newPercent: 85.0,
			},
		},
		allFiles: []string{"pkg/file1.go"},
		packages: []string{"pkg"},
	}

	rend := New(data, Options{
		ChangedFiles:           []string{"pkg/file1.go"},
		PackageThreshold:       5.0,
		PackageFileThreshold:   3.0,
		FileExclusionThreshold: 1.0,
	})

	jsonOutput := rend.JSON()
	require.NotEmpty(t, jsonOutput)

	// Check that JSON contains expected fields
	assert.Contains(t, jsonOutput, "ChangedFiles")
	assert.Contains(t, jsonOutput, "ChangedPackages")
	assert.Contains(t, jsonOutput, "PackageThreshold")
	assert.Contains(t, jsonOutput, "5")
	assert.Contains(t, jsonOutput, "pkg/file1.go")
}

func TestFilterExcludedFiles(t *testing.T) {
	files := []string{
		"pkg/file1.go",
		"pkg/file2_test.go",
		"pkg/file3.go",
		"helper.go",
	}

	// Note: filepath.Match doesn't support ** patterns, so we need exact paths
	patterns := []string{"pkg/file2_test.go", "helper.go"}

	filtered := filterExcludedFiles(files, patterns)

	assert.Len(t, filtered, 2)
	assert.Contains(t, filtered, "pkg/file1.go")
	assert.Contains(t, filtered, "pkg/file3.go")
	assert.NotContains(t, filtered, "pkg/file2_test.go")
	assert.NotContains(t, filtered, "helper.go")
}

func TestEmojiScore(t *testing.T) {
	tests := []struct {
		name        string
		oldPercent  float64
		newPercent  float64
		expectEmoji string
		expectDiff  string
	}{
		{
			name:        "large decrease",
			oldPercent:  80.0,
			newPercent:  20.0,
			expectEmoji: ":skull:",
			expectDiff:  "**-60.00%**",
		},
		{
			name:        "small decrease",
			oldPercent:  80.0,
			newPercent:  75.0,
			expectEmoji: ":thumbsdown:",
			expectDiff:  "**-5.00%**",
		},
		{
			name:        "no change",
			oldPercent:  80.0,
			newPercent:  80.0,
			expectEmoji: "",
			expectDiff:  "ø",
		},
		{
			name:        "small increase",
			oldPercent:  80.0,
			newPercent:  85.0,
			expectEmoji: ":thumbsup:",
			expectDiff:  "**+5.00%**",
		},
		{
			name:        "medium increase",
			oldPercent:  70.0,
			newPercent:  85.0,
			expectEmoji: ":tada:",
			expectDiff:  "**+15.00%**",
		},
		{
			name:        "large increase",
			oldPercent:  60.0,
			newPercent:  85.0,
			expectEmoji: ":star2:",
			expectDiff:  "**+25.00%**",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emoji, diff := emojiScore(tt.newPercent, tt.oldPercent)
			assert.Contains(t, emoji, tt.expectEmoji)
			assert.Equal(t, tt.expectDiff, diff)
		})
	}
}

func TestChangedPackages(t *testing.T) {
	files := []string{
		"pkg/subpkg/file1.go",
		"pkg/subpkg/file2.go",
		"pkg/file3.go",
		"other/file4.go",
	}

	packages := changedPackages(files)

	assert.Len(t, packages, 3)
	assert.Contains(t, packages, "pkg/subpkg")
	assert.Contains(t, packages, "pkg")
	assert.Contains(t, packages, "other")

	// Check that packages are sorted
	for i := 1; i < len(packages); i++ {
		assert.True(t, strings.Compare(packages[i-1], packages[i]) < 0)
	}
}
