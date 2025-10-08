package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fgrosse/go-coverage-report/internal/coveragechanges"
	"github.com/fgrosse/go-coverage-report/internal/renderer"
)

var usage = strings.TrimSpace(fmt.Sprintf(`
Usage: %s [OPTIONS] <OLD_COVERAGE_FILE> <NEW_COVERAGE_FILE> <CHANGED_FILES_FILE>

Parse the OLD_COVERAGE_FILE and NEW_COVERAGE_FILE and compare the coverage of the
files listed in CHANGED_FILES_FILE. The result is printed to stdout as a simple
Markdown table with emojis indicating the coverage change per package.

You can use the -root flag to add a prefix to all paths in the list of changed
files. This is useful to map the changed files (e.g., ["foo/my_file.go"] to their
coverage profile which uses the full package name to identify the files
(e.g., "github.com/fgrosse/example/foo/my_file.go"). Note that currently,
packages with a different name than their directory are not supported.

ARGUMENTS:
  OLD_COVERAGE_FILE   The path to the old coverage file in the format produced by go test -coverprofile
  NEW_COVERAGE_FILE   The path to the new coverage file in the same format as OLD_COVERAGE_FILE
  CHANGED_FILES_FILE  The path to the file containing the list of changed files encoded as JSON string array

OPTIONS:
`, filepath.Base(os.Args[0])))

type options struct {
	root                   string
	trim                   string
	format                 string
	excludeFile            string
	packageThreshold       float64
	packageFileThreshold   float64
	fileExclusionThreshold float64
}

func main() {
	log.SetFlags(0)

	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, usage)
		flag.PrintDefaults()
	}
	flag.String("root", "", "The import path of the tested repository to add as prefix to all paths of the changed files")
	flag.String("trim", "", "trim a prefix in the \"Impacted Packages\" column of the markdown report")
	flag.String("format", "markdown", "output format (currently only 'markdown' is supported)")
	flag.String("exclude-file", "", "path to file containing patterns for files to exclude from coverage report")
	flag.Float64("package-threshold", 0, "minimum coverage change percentage for packages to be included in Impacted Packages section")
	flag.Float64("package-file-threshold", 0, "minimum coverage change percentage for any file in a package to trigger package inclusion in Impacted Packages section")
	flag.Float64("file-exclusion-threshold", 0, "minimum coverage change percentage for files to be included in Changed files section")

	err := run(programArgs())
	if err != nil {
		log.Fatalln("ERROR:", err)
	}
}

func programArgs() (oldCov, newCov, changedFile string, opts options) {
	flag.Parse()

	args := flag.Args()
	if len(args) != 3 {
		if len(args) > 0 {
			log.Printf("ERROR: Expected exactly 3 arguments but got %d\n\n", len(args))
		}
		flag.Usage()
		os.Exit(1)
	}

	opts = options{
		root:                   flag.Lookup("root").Value.String(),
		trim:                   flag.Lookup("trim").Value.String(),
		format:                 flag.Lookup("format").Value.String(),
		excludeFile:            flag.Lookup("exclude-file").Value.String(),
		packageThreshold:       flag.Lookup("package-threshold").Value.(flag.Getter).Get().(float64),
		packageFileThreshold:   flag.Lookup("package-file-threshold").Value.(flag.Getter).Get().(float64),
		fileExclusionThreshold: flag.Lookup("file-exclusion-threshold").Value.(flag.Getter).Get().(float64),
	}

	return args[0], args[1], args[2], opts
}

// readExcludePatterns reads exclusion patterns from a file
func readExcludePatterns(filename string) ([]string, error) {
	if filename == "" {
		return nil, nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open exclude file %s: %w", filename, err)
	}
	defer file.Close()

	var patterns []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		patterns = append(patterns, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read exclude file %s: %w", filename, err)
	}

	return patterns, nil
}

// run uses the refactored architecture with separate packages
func run(oldCovPath, newCovPath, changedFilesPath string, opts options) error {
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
