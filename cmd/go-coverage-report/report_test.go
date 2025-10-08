package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReport_Markdown(t *testing.T) {
	oldCov, err := ParseCoverage("testdata/01-old-coverage.txt")
	require.NoError(t, err)

	newCov, err := ParseCoverage("testdata/01-new-coverage.txt")
	require.NoError(t, err)

	changedFiles, err := ParseChangedFiles("testdata/01-changed-files.json", "github.com/fgrosse/prioqueue")
	require.NoError(t, err)

	report := NewReport(oldCov, newCov, changedFiles, 0, 0, 0)
	actual := report.Markdown()

	expected := `### Merging this branch will **decrease** overall coverage

| Impacted Packages | Coverage Δ | :robot: |
|-------------------|------------|---------|
| github.com/fgrosse/prioqueue | 90.20% (**-9.80%**) | :thumbsdown: |
| github.com/fgrosse/prioqueue/foo/bar | 0.00% (ø) |  |

---

<details>

<summary>Coverage by file</summary>

### Changed files (no unit tests)

| Changed File | Coverage Δ | Total | Covered | Missed | :robot: |
|--------------|------------|-------|---------|--------|---------|
| github.com/fgrosse/prioqueue/foo/bar/baz.go | 0.00% (ø) | 0 | 0 | 0 |  |
| github.com/fgrosse/prioqueue/min_heap.go | 80.77% (**-19.23%**) | 52 (+2) | 42 (-8) | 10 (+10) | :skull:  |

_Please note that the "Total", "Covered", and "Missed" counts above refer to ***code statements*** instead of lines of code. The value in brackets refers to the test coverage of that file in the old version of the code._

</details>`
	assert.Equal(t, expected, actual)
}

func TestReport_Markdown_OnlyChangedUnitTests(t *testing.T) {
	oldCov, err := ParseCoverage("testdata/02-old-coverage.txt")
	require.NoError(t, err)

	newCov, err := ParseCoverage("testdata/02-new-coverage.txt")
	require.NoError(t, err)

	changedFiles, err := ParseChangedFiles("testdata/02-changed-files.json", "github.com/fgrosse/prioqueue")
	require.NoError(t, err)

	report := NewReport(oldCov, newCov, changedFiles, 0, 0, 0)
	actual := report.Markdown()

	expected := `### Merging this branch will **increase** overall coverage

| Impacted Packages | Coverage Δ | :robot: |
|-------------------|------------|---------|
| github.com/fgrosse/prioqueue | 99.02% (**+8.82%**) | :thumbsup: |

---

<details>

<summary>Coverage by file</summary>

### Changed unit test files

- github.com/fgrosse/prioqueue/min_heap_test.go

</details>`
	assert.Equal(t, expected, actual)
}

func TestReport_Filtering_WithThresholds(t *testing.T) {
	oldCov, err := ParseCoverage("testdata/threshold-old-coverage.txt")
	require.NoError(t, err)

	newCov, err := ParseCoverage("testdata/threshold-new-coverage.txt")
	require.NoError(t, err)

	changedFiles, err := ParseChangedFiles("testdata/threshold-changed-files.json", "github.com/fgrosse/prioqueue")
	require.NoError(t, err)

	// Test with package threshold that excludes packages with small changes
	report := NewReport(oldCov, newCov, changedFiles, 10.0, 1.0, 0)
	actual := report.Markdown()

	// Should only include the main package (90.20% coverage change is > 10.0%)
	// and exclude github.com/fgrosse/prioqueue/foo/bar (0.00% change is < 10.0%)
	expected := `### Merging this branch will **decrease** overall coverage

| Impacted Packages | Coverage Δ | :robot: |
|-------------------|------------|---------|
| github.com/fgrosse/prioqueue | 90.20% (**-9.80%**) | :thumbsdown: |

---

<details>

<summary>Coverage by file</summary>

### Changed files (no unit tests)

| Changed File | Coverage Δ | Total | Covered | Missed | :robot: |
|--------------|------------|-------|---------|--------|---------|
| github.com/fgrosse/prioqueue/foo/bar/baz.go | 0.00% (ø) | 0 | 0 | 0 |  |
| github.com/fgrosse/prioqueue/min_heap.go | 80.77% (**-19.23%**) | 52 (+2) | 42 (-8) | 10 (+10) | :skull:  |

_Please note that the "Total", "Covered", and "Missed" counts above refer to ***code statements*** instead of lines of code. The value in brackets refers to the test coverage of that file in the old version of the code._

</details>`
	assert.Equal(t, expected, actual)
}

func TestReport_Filtering_FileExclusionThreshold(t *testing.T) {
	oldCov, err := ParseCoverage("testdata/threshold-old-coverage.txt")
	require.NoError(t, err)

	newCov, err := ParseCoverage("testdata/threshold-new-coverage.txt")
	require.NoError(t, err)

	changedFiles, err := ParseChangedFiles("testdata/threshold-changed-files.json", "github.com/fgrosse/prioqueue")
	require.NoError(t, err)

	// Test with file exclusion threshold that excludes files with small changes
	report := NewReport(oldCov, newCov, changedFiles, 0, 0, 10.0)
	actual := report.Markdown()

	// Should exclude github.com/fgrosse/prioqueue/foo/bar/baz.go (0.00% change is < 10.0%)
	// and include github.com/fgrosse/prioqueue/min_heap.go (19.23% change is > 10.0%)
	expected := `### Merging this branch will **decrease** overall coverage

| Impacted Packages | Coverage Δ | :robot: |
|-------------------|------------|---------|
| github.com/fgrosse/prioqueue | 90.20% (**-9.80%**) | :thumbsdown: |
| github.com/fgrosse/prioqueue/foo/bar | 0.00% (ø) |  |

---

<details>

<summary>Coverage by file</summary>

### Changed files (no unit tests)

| Changed File | Coverage Δ | Total | Covered | Missed | :robot: |
|--------------|------------|-------|---------|--------|---------|
| github.com/fgrosse/prioqueue/min_heap.go | 80.77% (**-19.23%**) | 52 (+2) | 42 (-8) | 10 (+10) | :skull:  |

_Please note that the "Total", "Covered", and "Missed" counts above refer to ***code statements*** instead of lines of code. The value in brackets refers to the test coverage of that file in the old version of the code._

</details>`
	assert.Equal(t, expected, actual)
}

func TestReport_Filtering_PackageFileThreshold(t *testing.T) {
	oldCov, err := ParseCoverage("testdata/threshold-old-coverage.txt")
	require.NoError(t, err)

	newCov, err := ParseCoverage("testdata/threshold-new-coverage.txt")
	require.NoError(t, err)

	changedFiles, err := ParseChangedFiles("testdata/threshold-changed-files.json", "github.com/fgrosse/prioqueue")
	require.NoError(t, err)

	// Test with package file threshold that includes packages based on file changes
	report := NewReport(oldCov, newCov, changedFiles, 0, 15.0, 0)
	actual := report.Markdown()

	// Should include github.com/fgrosse/prioqueue because min_heap.go has 19.23% change (> 15.0%)
	// Should include github.com/fgrosse/prioqueue/foo/bar because baz.go has 0% change but threshold is 0
	expected := `### Merging this branch will **decrease** overall coverage

| Impacted Packages | Coverage Δ | :robot: |
|-------------------|------------|---------|
| github.com/fgrosse/prioqueue | 90.20% (**-9.80%**) | :thumbsdown: |
| github.com/fgrosse/prioqueue/foo/bar | 0.00% (ø) |  |

---

<details>

<summary>Coverage by file</summary>

### Changed files (no unit tests)

| Changed File | Coverage Δ | Total | Covered | Missed | :robot: |
|--------------|------------|-------|---------|--------|---------|
| github.com/fgrosse/prioqueue/foo/bar/baz.go | 0.00% (ø) | 0 | 0 | 0 |  |
| github.com/fgrosse/prioqueue/min_heap.go | 80.77% (**-19.23%**) | 52 (+2) | 42 (-8) | 10 (+10) | :skull:  |

_Please note that the "Total", "Covered", and "Missed" counts above refer to ***code statements*** instead of lines of code. The value in brackets refers to the test coverage of that file in the old version of the code._

</details>`
	assert.Equal(t, expected, actual)
}
