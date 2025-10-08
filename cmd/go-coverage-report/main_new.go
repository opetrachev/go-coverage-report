package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fgrosse/go-coverage-report/internal/coveragechanges"
	"github.com/fgrosse/go-coverage-report/internal/renderer"
)

// runNew uses the new refactored architecture with separate packages
func runNew(oldCovPath, newCovPath, changedFilesPath string, opts options) error {
	// Read exclude patterns if specified
	excludePatterns, err := readExcludePatterns(opts.excludeFile)
	if err != nil {
		return fmt.Errorf("failed to read exclude patterns: %w", err)
	}

	// Parse coverage files without exclusions
	oldCov, err := ParseCoverage(oldCovPath)
	if err != nil {
		return fmt.Errorf("failed to parse old coverage: %w", err)
	}

	newCov, err := ParseCoverage(newCovPath)
	if err != nil {
		return fmt.Errorf("failed to parse new coverage: %w", err)
	}

	// Parse changed files without exclusions
	changedFiles, err := ParseChangedFiles(changedFilesPath, opts.root)
	if err != nil {
		return fmt.Errorf("failed to load changed files: %w", err)
	}

	if len(changedFiles) == 0 {
		log.Println("Skipping report since there are no changed files")
		return nil
	}

	// Create coverage changes using the new package
	oldCovAdapter := NewCoverageAdapter(oldCov)
	newCovAdapter := NewCoverageAdapter(newCov)
	changes := coveragechanges.New(oldCovAdapter, newCovAdapter)

	// Create renderer with the new package
	changesAdapter := NewCoverageChangesAdapter(changes)
	rend := renderer.New(changesAdapter, renderer.Options{
		ChangedFiles:           changedFiles,
		PackageThreshold:       opts.packageThreshold,
		PackageFileThreshold:   opts.packageFileThreshold,
		FileExclusionThreshold: opts.fileExclusionThreshold,
		ExcludePatterns:        excludePatterns,
		TrimPrefix:             opts.trim,
	})

	// Generate output based on format
	switch strings.ToLower(opts.format) {
	case "markdown":
		fmt.Fprintln(os.Stdout, rend.Markdown())
	case "json":
		fmt.Fprintln(os.Stdout, rend.JSON())
	default:
		return fmt.Errorf("unsupported format: %q", opts.format)
	}

	return nil
}
