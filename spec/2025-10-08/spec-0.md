# Refactoring: split changes calculation and  output generation into separate packages

Currently, for both the report creation and for the generation of output files, a single `Report` type 
and its methods are used. `main.go` uses `NewReport()` to create the report struct, and then calls `report.Markdown()`
or `report.JSON()` to print it.

The logic of changes calculation and report output generation intertwine in the same methods in the `Report`.

This is a problem, because it is impossible to imdependently implement and test changes in calculation logic and output generation logic. They are tightly coupled with each other.

We need to refactor it. As a result of the refactoring, 2 new packages should be created instead of the existing `Report` package:

1. First one will be used for receiving 2 coverage profiles and building a struct, reprecenting coverage changes between these coverage profiles. The purpose of this package is to **calculate** differences in coverage.

2. Second one will receive the generated struct and only print it the requested form (Markdown or JSON), without making any calculations. It also should support filtering by thresholds, so thresholds logic and file exclusion logic goes into this package. The purpose of this package is to print the report, potentially after applying filtering.

## What do you need to do

To clarify the necessary details, you will ask me questions in chat in russian, and I will provide you with answers. 

Ask me from time to time if we need anything else to discuss, or suggest your questions. Don't start the implementation until we discuss all the necessary implementation aspects.

When we are done with the discussion, you will save the information, important for the implementation, here under the Discussion section, in english.

## Discussion

### Key Decisions from Discussion

#### 1. Package Structure

**First Package (Changes Calculation):**
- Package name: `coveragechanges` (located in `internal/coveragechanges`)
- Main type: `CoverageChanges`
- Data structure: List of all files with their oldCov and newCov, plus calculated changes
- All coverage percentage changes should be pre-calculated
- Initially, keep it simple - just files and their changes, no packages in the data structure (to avoid complexity)

**Second Package (Output Rendering):**
- Package name: `renderer` (located in `internal/renderer`)
- Purpose: Format output with filtering applied

#### 2. Filtering and Exclusions

**At Parsing Stage:**
- No exclusions during coverage file parsing
- No exclusions when reading changed files list
- `ParseCoverage()` and `ParseChangedFiles()` signatures changed to remove `excludePatterns` parameter

**At Calculation Stage:**
- `CoverageChanges` should contain ALL files from both coverage profiles
- No filtering at this stage

**At Rendering Stage:**
- ALL filtering happens in the renderer package:
  - File exclusion by patterns (`excludePatterns`)
  - Package threshold filtering (`packageThreshold`)
  - Package file threshold filtering (`packageFileThreshold`)
  - File exclusion threshold filtering (`fileExclusionThreshold`)
- `TrimPrefix` logic also belongs to renderer

#### 3. Interface Design

**Renderer Interface:**
- Renderer package declares an interface for working with coverage data
- The interface methods should provide access to:
  - `oldPercent`, `newPercent`, `delta` for files
  - Total, Covered, Missed statements (not just percentages - these are shown in reports)
  - Package-level aggregated data

**Dependency Direction:**
- Renderer package does NOT import coveragechanges package
- Renderer declares interface, coveragechanges implements it via adapter
- Integration happens in `main.go`

#### 4. Backward Compatibility

**API Compatibility:**
- Current API must remain the same
- Old `report.go` should NOT be touched
- Old tests should continue to work

**Testing:**
- New packages need tests, but happy path scenarios are sufficient
- Old report tests remain unchanged

#### 5. Package Location

Packages should be created in `internal/` directory following Go conventions:
- `internal/coveragechanges/`
- `internal/renderer/`

#### 6. Changed Files

- `changedFiles` list is needed by the renderer to know which files to display
- This list comes from user input (JSON file with changed files)
- The list should be passed to renderer, not to coveragechanges