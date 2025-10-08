package coveragechanges

import (
	"math"
	"path/filepath"
	"sort"
)

// FileChange represents coverage change for a single file
type FileChange struct {
	FileName    string
	OldCoverage *FileCoverage
	NewCoverage *FileCoverage
}

// FileCoverage represents coverage metrics for a file
type FileCoverage struct {
	Percent     float64
	TotalStmt   int64
	CoveredStmt int64
	MissedStmt  int64
}

// CoverageChanges represents all coverage changes between two coverage profiles
type CoverageChanges struct {
	files map[string]*FileChange
}

// Coverage represents a parsed coverage profile
type Coverage interface {
	GetFiles() map[string]FileProfile
}

// FileProfile represents coverage data for a single file in the coverage profile
type FileProfile interface {
	CoveragePercent() float64
	GetTotal() int64
	GetCovered() int64
	GetMissed() int64
}

// New creates a new CoverageChanges from old and new coverage profiles
func New(oldCov, newCov Coverage) *CoverageChanges {
	changes := &CoverageChanges{
		files: make(map[string]*FileChange),
	}

	allFiles := make(map[string]bool)

	// Get files maps once before loops
	oldFiles := oldCov.GetFiles()
	newFiles := newCov.GetFiles()

	// Collect all files from both coverages
	for file := range oldFiles {
		allFiles[file] = true
	}
	for file := range newFiles {
		allFiles[file] = true
	}

	// Create FileChange for each file
	for file := range allFiles {
		oldProfile := oldFiles[file]
		newProfile := newFiles[file]

		fileChange := &FileChange{
			FileName: file,
		}

		if oldProfile != nil {
			fileChange.OldCoverage = &FileCoverage{
				Percent:     oldProfile.CoveragePercent(),
				TotalStmt:   oldProfile.GetTotal(),
				CoveredStmt: oldProfile.GetCovered(),
				MissedStmt:  oldProfile.GetMissed(),
			}
		}

		if newProfile != nil {
			fileChange.NewCoverage = &FileCoverage{
				Percent:     newProfile.CoveragePercent(),
				TotalStmt:   newProfile.GetTotal(),
				CoveredStmt: newProfile.GetCovered(),
				MissedStmt:  newProfile.GetMissed(),
			}
		}

		changes.files[file] = fileChange
	}

	return changes
}

// GetFileChange returns the coverage change for a specific file
func (c *CoverageChanges) GetFileChange(file string) *FileChange {
	return c.files[file]
}

// GetAllFiles returns all files with coverage data
func (c *CoverageChanges) GetAllFiles() []string {
	files := make([]string, 0, len(c.files))
	for file := range c.files {
		files = append(files, file)
	}
	sort.Strings(files)
	return files
}

// GetPackageChange returns aggregated coverage change for a package
func (c *CoverageChanges) GetPackageChange(pkg string) *FileChange {
	var oldTotal, oldCovered, newTotal, newCovered int64

	for file, change := range c.files {
		if filepath.Dir(file) != pkg {
			continue
		}

		if change.OldCoverage != nil {
			oldTotal += change.OldCoverage.TotalStmt
			oldCovered += change.OldCoverage.CoveredStmt
		}

		if change.NewCoverage != nil {
			newTotal += change.NewCoverage.TotalStmt
			newCovered += change.NewCoverage.CoveredStmt
		}
	}

	pkgChange := &FileChange{
		FileName: pkg,
	}

	if oldTotal > 0 {
		pkgChange.OldCoverage = &FileCoverage{
			Percent:     float64(oldCovered) / float64(oldTotal) * 100,
			TotalStmt:   oldTotal,
			CoveredStmt: oldCovered,
			MissedStmt:  oldTotal - oldCovered,
		}
	}

	if newTotal > 0 {
		pkgChange.NewCoverage = &FileCoverage{
			Percent:     float64(newCovered) / float64(newTotal) * 100,
			TotalStmt:   newTotal,
			CoveredStmt: newCovered,
			MissedStmt:  newTotal - newCovered,
		}
	}

	return pkgChange
}

// GetPackages returns all unique packages from all files
func (c *CoverageChanges) GetPackages() []string {
	packages := make(map[string]bool)

	for file := range c.files {
		pkg := filepath.Dir(file)
		packages[pkg] = true
	}

	result := make([]string, 0, len(packages))
	for pkg := range packages {
		result = append(result, pkg)
	}
	sort.Strings(result)

	return result
}

// CoverageChange calculates the absolute coverage change between old and new percentages
func CoverageChange(oldPercent, newPercent float64) float64 {
	return math.Abs(newPercent - oldPercent)
}

// GetOldPercent returns old coverage percentage for a file, or 0 if not present
func (fc *FileChange) GetOldPercent() float64 {
	if fc.OldCoverage == nil {
		return 0
	}
	return fc.OldCoverage.Percent
}

// GetNewPercent returns new coverage percentage for a file, or 0 if not present
func (fc *FileChange) GetNewPercent() float64 {
	if fc.NewCoverage == nil {
		return 0
	}
	return fc.NewCoverage.Percent
}

// GetDelta returns the delta between new and old coverage
func (fc *FileChange) GetDelta() float64 {
	return fc.GetNewPercent() - fc.GetOldPercent()
}
