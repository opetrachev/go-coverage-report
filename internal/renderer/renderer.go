package renderer

import (
	"encoding/json"
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strings"
)

// CoverageData is an interface that provides coverage change data for rendering
type CoverageData interface {
	// GetFileChange returns the coverage change for a specific file
	GetFileChange(file string) FileChange
	// GetPackageChange returns aggregated coverage change for a package
	GetPackageChange(pkg string) FileChange
	// GetAllFiles returns all files with coverage data
	GetAllFiles() []string
	// GetPackages returns all unique packages
	GetPackages() []string
}

// FileChange represents coverage change for a single file or package
type FileChange interface {
	GetOldPercent() float64
	GetNewPercent() float64
	GetDelta() float64
	GetOldCoverage() FileCoverage
	GetNewCoverage() FileCoverage
}

// FileCoverage represents coverage metrics
type FileCoverage interface {
	GetPercent() float64
	GetTotalStmt() int64
	GetCoveredStmt() int64
	GetMissedStmt() int64
}

// Renderer renders coverage reports in different formats
type Renderer struct {
	data                   CoverageData
	changedFiles           []string
	changedPackages        []string
	packageThreshold       float64
	packageFileThreshold   float64
	fileExclusionThreshold float64
	excludePatterns        []string
	trimPrefixValue        string
}

// Options contains configuration for the renderer
type Options struct {
	ChangedFiles           []string
	PackageThreshold       float64
	PackageFileThreshold   float64
	FileExclusionThreshold float64
	ExcludePatterns        []string
	TrimPrefix             string
}

// New creates a new Renderer with the given coverage data and options
func New(data CoverageData, opts Options) *Renderer {
	changedFiles := filterExcludedFiles(opts.ChangedFiles, opts.ExcludePatterns)
	sort.Strings(changedFiles)

	return &Renderer{
		data:                   data,
		changedFiles:           changedFiles,
		changedPackages:        changedPackages(changedFiles),
		packageThreshold:       opts.PackageThreshold,
		packageFileThreshold:   opts.PackageFileThreshold,
		fileExclusionThreshold: opts.FileExclusionThreshold,
		excludePatterns:        opts.ExcludePatterns,
		trimPrefixValue:        opts.TrimPrefix,
	}
}

// changedPackages extracts unique packages from changed files
func changedPackages(changedFiles []string) []string {
	packages := map[string]bool{}
	for _, file := range changedFiles {
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

// filterExcludedFiles filters out files matching exclusion patterns
func filterExcludedFiles(files []string, excludePatterns []string) []string {
	if len(excludePatterns) == 0 {
		return files
	}

	var filtered []string
	for _, file := range files {
		shouldExclude := false
		for _, pattern := range excludePatterns {
			if matched, _ := filepath.Match(pattern, file); matched {
				shouldExclude = true
				break
			}
		}
		if !shouldExclude {
			filtered = append(filtered, file)
		}
	}

	return filtered
}

// coverageChange calculates the absolute coverage change
func coverageChange(oldPercent, newPercent float64) float64 {
	return math.Abs(newPercent - oldPercent)
}

// shouldIncludePackage determines if a package should be included in the output
func (r *Renderer) shouldIncludePackage(pkg string) bool {
	pkgChange := r.data.GetPackageChange(pkg)
	packageCoverageChange := coverageChange(pkgChange.GetOldPercent(), pkgChange.GetNewPercent())

	// Check package threshold
	if packageCoverageChange >= r.packageThreshold {
		return true
	}

	// Check if any file in the package meets the package file threshold
	for _, file := range r.changedFiles {
		if filepath.Dir(file) == pkg {
			fileChange := r.data.GetFileChange(file)
			fileCoverageChange := coverageChange(fileChange.GetOldPercent(), fileChange.GetNewPercent())
			if fileCoverageChange >= r.packageFileThreshold {
				return true
			}
		}
	}

	return false
}

// shouldIncludeFile determines if a file should be included in the output
func (r *Renderer) shouldIncludeFile(file string) bool {
	fileChange := r.data.GetFileChange(file)
	fileCoverageChange := coverageChange(fileChange.GetOldPercent(), fileChange.GetNewPercent())
	return fileCoverageChange >= r.fileExclusionThreshold
}

// filteredChangedPackages returns packages that meet the threshold criteria
func (r *Renderer) filteredChangedPackages() []string {
	var filtered []string
	for _, pkg := range r.changedPackages {
		if r.shouldIncludePackage(pkg) {
			filtered = append(filtered, pkg)
		}
	}
	return filtered
}

// filteredChangedFiles returns files that meet the exclusion threshold criteria
func (r *Renderer) filteredChangedFiles() []string {
	var filtered []string
	for _, file := range r.changedFiles {
		if r.shouldIncludeFile(file) {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

// trimPrefix removes a prefix from a string
func (r *Renderer) trimPrefix(name string) string {
	if r.trimPrefixValue == "" {
		return name
	}

	trimmed := strings.TrimPrefix(name, r.trimPrefixValue)
	trimmed = strings.TrimPrefix(trimmed, "/")
	if trimmed == "" {
		trimmed = "."
	}

	return trimmed
}

// Title generates the report title based on coverage changes
func (r *Renderer) Title() string {
	var numDecrease, numIncrease int
	filteredPackages := r.filteredChangedPackages()

	for _, pkg := range filteredPackages {
		pkgChange := r.data.GetPackageChange(pkg)
		oldPercent := pkgChange.GetOldPercent()
		newPercent := pkgChange.GetNewPercent()

		newP := round(newPercent, 2)
		oldP := round(oldPercent, 2)

		switch {
		case newP > oldP:
			numIncrease++
		case newP < oldP:
			numDecrease++
		}
	}

	switch {
	case numIncrease == 0 && numDecrease == 0:
		return fmt.Sprintln("### Merging this branch will **not change** overall coverage")
	case numIncrease > 0 && numDecrease == 0:
		return fmt.Sprintln("### Merging this branch will **increase** overall coverage")
	case numIncrease == 0 && numDecrease > 0:
		return fmt.Sprintln("### Merging this branch will **decrease** overall coverage")
	default:
		return fmt.Sprintf("### Merging this branch changes the coverage (%d decrease, %d increase)\n", numDecrease, numIncrease)
	}
}

// Markdown generates a Markdown report
func (r *Renderer) Markdown() string {
	report := new(strings.Builder)

	fmt.Fprintln(report, r.Title())
	fmt.Fprintln(report, "| Impacted Packages | Coverage Δ | :robot: |")
	fmt.Fprintln(report, "|-------------------|------------|---------|")

	filteredPackages := r.filteredChangedPackages()
	for _, pkg := range filteredPackages {
		pkgChange := r.data.GetPackageChange(pkg)
		oldPercent := pkgChange.GetOldPercent()
		newPercent := pkgChange.GetNewPercent()

		emoji, diffStr := emojiScore(newPercent, oldPercent)
		fmt.Fprintf(report, "| %s | %.2f%% (%s) | %s |\n",
			r.trimPrefix(pkg),
			newPercent,
			diffStr,
			emoji,
		)
	}

	report.WriteString("\n")
	r.addDetails(report)

	return report.String()
}

// addDetails adds the detailed coverage by file section
func (r *Renderer) addDetails(report *strings.Builder) {
	fmt.Fprintln(report, "---")
	fmt.Fprintln(report)
	fmt.Fprintln(report, "<details>")
	fmt.Fprintln(report)

	fmt.Fprintln(report, "<summary>Coverage by file</summary>")
	fmt.Fprintln(report)

	var codeFiles, unitTestFiles []string
	filteredFiles := r.filteredChangedFiles()
	for _, f := range filteredFiles {
		if strings.HasSuffix(f, "_test.go") {
			unitTestFiles = append(unitTestFiles, f)
		} else {
			codeFiles = append(codeFiles, f)
		}
	}

	if len(codeFiles) > 0 {
		r.addCodeFileDetails(report, codeFiles)
	}
	if len(unitTestFiles) > 0 {
		r.addTestFileDetails(report, unitTestFiles)
	}

	fmt.Fprint(report, "</details>")
}

// addCodeFileDetails adds the code files section
func (r *Renderer) addCodeFileDetails(report *strings.Builder, files []string) {
	fmt.Fprintln(report, "### Changed files (no unit tests)")
	fmt.Fprintln(report)
	fmt.Fprintln(report, "| Changed File | Coverage Δ | Total | Covered | Missed | :robot: |")
	fmt.Fprintln(report, "|--------------|------------|-------|---------|--------|---------|")

	for _, name := range files {
		fileChange := r.data.GetFileChange(name)
		oldCov := fileChange.GetOldCoverage()
		newCov := fileChange.GetNewCoverage()

		oldPercent := fileChange.GetOldPercent()
		newPercent := fileChange.GetNewPercent()

		valueWithDelta := func(oldVal, newVal int64) string {
			diff := oldVal - newVal
			switch {
			case diff < 0:
				return fmt.Sprintf("%d (+%d)", newVal, -diff)
			case diff > 0:
				return fmt.Sprintf("%d (-%d)", newVal, diff)
			default:
				return fmt.Sprintf("%d", newVal)
			}
		}

		emoji, diffStr := emojiScore(newPercent, oldPercent)
		fmt.Fprintf(report, "| %s | %.2f%% (%s) | %s | %s | %s | %s |\n",
			r.trimPrefix(name),
			newPercent, diffStr,
			valueWithDelta(oldCov.GetTotalStmt(), newCov.GetTotalStmt()),
			valueWithDelta(oldCov.GetCoveredStmt(), newCov.GetCoveredStmt()),
			valueWithDelta(oldCov.GetMissedStmt(), newCov.GetMissedStmt()),
			emoji,
		)
	}

	fmt.Fprintln(report)
	fmt.Fprintln(report, `_Please note that the "Total", "Covered", and "Missed" counts `+
		"above refer to ***code statements*** instead of lines of code. The value in brackets "+
		"refers to the test coverage of that file in the old version of the code._")
	fmt.Fprintln(report)
}

// addTestFileDetails adds the test files section
func (r *Renderer) addTestFileDetails(report *strings.Builder, files []string) {
	fmt.Fprintln(report, "### Changed unit test files")
	fmt.Fprintln(report)

	for _, name := range files {
		fmt.Fprintf(report, "- %s\n", r.trimPrefix(name))
	}

	fmt.Fprintln(report)
}

// JSON generates a JSON report
func (r *Renderer) JSON() string {
	// Create a structure that matches the old Report structure for JSON output
	type jsonReport struct {
		Old                    map[string]interface{} `json:"Old"`
		New                    map[string]interface{} `json:"New"`
		ChangedFiles           []string               `json:"ChangedFiles"`
		ChangedPackages        []string               `json:"ChangedPackages"`
		PackageThreshold       float64                `json:"PackageThreshold"`
		PackageFileThreshold   float64                `json:"PackageFileThreshold"`
		FileExclusionThreshold float64                `json:"FileExclusionThreshold"`
	}

	jr := jsonReport{
		Old:                    make(map[string]interface{}),
		New:                    make(map[string]interface{}),
		ChangedFiles:           r.changedFiles,
		ChangedPackages:        r.changedPackages,
		PackageThreshold:       r.packageThreshold,
		PackageFileThreshold:   r.packageFileThreshold,
		FileExclusionThreshold: r.fileExclusionThreshold,
	}

	data, err := json.MarshalIndent(jr, "", "    ")
	if err != nil {
		panic(err) // should never happen
	}

	return string(data)
}

// round rounds a value to a specified number of decimal places
func round(val float64, places int) float64 {
	if val == 0 {
		return 0
	}

	pow := math.Pow10(places)
	digit := math.Round(pow * val)
	return digit / pow
}

// emojiScore returns an emoji and diff string based on coverage change
func emojiScore(newPercent, oldPercent float64) (emoji, diffStr string) {
	diff := newPercent - oldPercent
	switch {
	case diff < -50:
		emoji = strings.Repeat(":skull: ", 5)
		diffStr = fmt.Sprintf("**%+.2f%%**", diff)
	case diff < -10:
		emoji = strings.Repeat(":skull: ", int(-diff/10))
		diffStr = fmt.Sprintf("**%+.2f%%**", diff)
	case diff < 0:
		emoji = ":thumbsdown:"
		diffStr = fmt.Sprintf("**%+.2f%%**", diff)
	case diff == 0:
		emoji = ""
		diffStr = "ø"
	case diff > 20:
		emoji = ":star2:"
		diffStr = fmt.Sprintf("**%+.2f%%**", diff)
	case diff > 10:
		emoji = ":tada:"
		diffStr = fmt.Sprintf("**%+.2f%%**", diff)
	case diff > 0:
		emoji = ":thumbsup:"
		diffStr = fmt.Sprintf("**%+.2f%%**", diff)
	}

	return emoji, diffStr
}
