package main

import (
	"github.com/fgrosse/go-coverage-report/internal/coveragechanges"
	"github.com/fgrosse/go-coverage-report/internal/renderer"
)

// CoverageAdapter adapts Coverage to implement coveragechanges.Coverage interface
type CoverageAdapter struct {
	cov *Coverage
}

func NewCoverageAdapter(cov *Coverage) *CoverageAdapter {
	return &CoverageAdapter{cov: cov}
}

func (ca *CoverageAdapter) GetFiles() map[string]coveragechanges.FileProfile {
	result := make(map[string]coveragechanges.FileProfile)
	for name, profile := range ca.cov.Files {
		result[name] = &ProfileAdapter{profile: profile}
	}
	return result
}

// ProfileAdapter adapts Profile to implement coveragechanges.FileProfile interface
type ProfileAdapter struct {
	profile *Profile
}

func (pa *ProfileAdapter) CoveragePercent() float64 {
	return pa.profile.CoveragePercent()
}

func (pa *ProfileAdapter) GetTotal() int64 {
	return pa.profile.GetTotal()
}

func (pa *ProfileAdapter) GetCovered() int64 {
	return pa.profile.GetCovered()
}

func (pa *ProfileAdapter) GetMissed() int64 {
	return pa.profile.GetMissed()
}

// CoverageChangesAdapter adapts CoverageChanges to implement renderer.CoverageData interface
type CoverageChangesAdapter struct {
	changes *coveragechanges.CoverageChanges
}

func NewCoverageChangesAdapter(changes *coveragechanges.CoverageChanges) *CoverageChangesAdapter {
	return &CoverageChangesAdapter{changes: changes}
}

func (cca *CoverageChangesAdapter) GetFileChange(file string) renderer.FileChange {
	fc := cca.changes.GetFileChange(file)
	if fc == nil {
		return nil
	}
	return &FileChangeAdapter{fc: fc}
}

func (cca *CoverageChangesAdapter) GetPackageChange(pkg string) renderer.FileChange {
	fc := cca.changes.GetPackageChange(pkg)
	if fc == nil {
		return nil
	}
	return &FileChangeAdapter{fc: fc}
}

func (cca *CoverageChangesAdapter) GetAllFiles() []string {
	return cca.changes.GetAllFiles()
}

func (cca *CoverageChangesAdapter) GetPackages() []string {
	return cca.changes.GetPackages()
}

// FileChangeAdapter adapts coveragechanges.FileChange to implement renderer.FileChange interface
type FileChangeAdapter struct {
	fc *coveragechanges.FileChange
}

func (fca *FileChangeAdapter) GetOldPercent() float64 {
	return fca.fc.GetOldPercent()
}

func (fca *FileChangeAdapter) GetNewPercent() float64 {
	return fca.fc.GetNewPercent()
}

func (fca *FileChangeAdapter) GetDelta() float64 {
	return fca.fc.GetDelta()
}

func (fca *FileChangeAdapter) GetOldCoverage() renderer.FileCoverage {
	if fca.fc.OldCoverage == nil {
		return &EmptyFileCoverage{}
	}
	return &FileCoverageAdapter{fc: fca.fc.OldCoverage}
}

func (fca *FileChangeAdapter) GetNewCoverage() renderer.FileCoverage {
	if fca.fc.NewCoverage == nil {
		return &EmptyFileCoverage{}
	}
	return &FileCoverageAdapter{fc: fca.fc.NewCoverage}
}

// FileCoverageAdapter adapts coveragechanges.FileCoverage to implement renderer.FileCoverage interface
type FileCoverageAdapter struct {
	fc *coveragechanges.FileCoverage
}

func (fca *FileCoverageAdapter) GetPercent() float64 {
	return fca.fc.Percent
}

func (fca *FileCoverageAdapter) GetTotalStmt() int64 {
	return fca.fc.TotalStmt
}

func (fca *FileCoverageAdapter) GetCoveredStmt() int64 {
	return fca.fc.CoveredStmt
}

func (fca *FileCoverageAdapter) GetMissedStmt() int64 {
	return fca.fc.MissedStmt
}

// EmptyFileCoverage represents an empty coverage (when file doesn't exist in one of the profiles)
type EmptyFileCoverage struct{}

func (efc *EmptyFileCoverage) GetPercent() float64 {
	return 0
}

func (efc *EmptyFileCoverage) GetTotalStmt() int64 {
	return 0
}

func (efc *EmptyFileCoverage) GetCoveredStmt() int64 {
	return 0
}

func (efc *EmptyFileCoverage) GetMissedStmt() int64 {
	return 0
}
