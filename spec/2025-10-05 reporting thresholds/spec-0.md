# Reportng thresholds feature

I want to create a feature to exclude some files and packages from the report, based on a threshold of how much code coverage has changed for a file of a package.

There will be new command line switches:

 * Package threshold, float, represents coverage in precents. Example: 0.5. Default 0.
    * Affects packages, shown in the "Impacted Packages" section.
 * File threshold, float, also represents coverage in percents. Example: 5. Default 0.
    * Also affects packages, shown in the "Impacted Packages" section.
 * File exclusion threshold, float, represents coverage in precents. Example: 0.05. Default 0.
    * Affects files, shown in the "Changed files" section.

## Requirements

### Package threshold and File threshold

 1. Even if these thresholds are set, all files and packages are still included in code coverage changes calculations. Thresholds affect only report generation, not calculations.
 2. The logic is not recursive. Current package is not affected by files in its sub-packages.
 3. In the report, only the "Impacted Packages" section is affected. The "Changed files" section should still represent all files, as normal.
 4. If a package has coverage changes are equal or greater than the PACKAGE_THRESHOLD, it is included in the "Impacted Packages" section.
 5. If one of the package's files has coverage changes equal or greater than the FILE_THRESHOLD, the package is also included into the "Impacted Packages" section.
 6. Otherwise, the package is not included into the "Impacted Packages" section.
 7. In any way, the "Changed files" section is not impacted and still represents all files.

### File exclusion threshold

 1. Even if the threshold is set, all files and packages are still included in code coverage changes calculations. The treshold affects only report generation, not calculations.
 2. The file exclusion threshold affects the "Changed files" section.
 3. If a file exclusion threshold is set, and file's code coverage changes are strictly less than the threshold, the file is not included into the "Changed files" section.
 4. Otherwise, everything works as usual and it is included.
 5. To protect from flag conflicts, the File exclusion threshold can not be greater than the File threshold.

## What do you need to do

To clarify the necessary details, you will ask me questions in chat in russian, and I will provide you with answers. 

The new feature introduces new terms for the Uniquity language of the application. If you see better naming, make a suggestion.

Ask me from time to time if we need anything else to discuss or suggest your questions, don't start the implementation until we discuss all the necessary implementation aspects.

When we are done with the discussion, you will save the information, important for the implementation, here under the Discussion section, in english.

Also, you will save the final terminology glossary here under the "Ubiquity language (Application/Domain terminology)" section, also in english.

## Discussion

### Implementation Details

1. **Threshold Units**: All thresholds are specified in percentage points. Examples: 0.5 = 0.5%, 5 = 5%

2. **Command Line Parameters**:
   - `--package-threshold`: Float, coverage percentage for packages. Default: 0
   - `--package-file-threshold`: Float, coverage percentage for files within packages. Default: 0  
   - `--file-exclusion-threshold`: Float, coverage percentage for file exclusion. Default: 0

3. **Package Filtering Logic**:
   - A package is included in "Impacted Packages" section if:
     - `|package_coverage_change| >= package-threshold` OR
     - `|any_file_coverage_change_in_package| >= package-file-threshold`
   - Uses absolute values for comparison (handles negative coverage changes)
   - Logic is OR-based: either condition can trigger inclusion

4. **File Filtering Logic**:
   - A file is included in "Changed files" section if:
     - `|file_coverage_change| >= file-exclusion-threshold`
   - Uses absolute values for comparison
   - Files with zero coverage change are included by default (since default threshold is 0)

5. **Validation**:
   - `file-exclusion-threshold <= package-file-threshold` (protection against flag conflicts)
   - No range validation required - any parseable number is accepted

6. **Backward Compatibility**:
   - All thresholds default to 0, maintaining current behavior when flags are not specified
   - All files and packages included by default (existing functionality preserved)

7. **Coverage Change Calculation**:
   - Uses existing coverage change calculations already present in the codebase
   - No changes to calculation logic, only filtering based on computed values

## Ubiquity language (Application/Domain terminology)

### Core Terms

- **Package Threshold**: Minimum absolute coverage change percentage required for a package to be included in "Impacted Packages" section
- **Package File Threshold**: Minimum absolute coverage change percentage for any file within a package to trigger package inclusion in "Impacted Packages" section  
- **File Exclusion Threshold**: Minimum absolute coverage change percentage required for a file to be included in "Changed files" section
- **Coverage Change**: Absolute difference between old and new coverage percentages (handles both positive and negative changes)
- **Impacted Packages**: Packages section in report that shows coverage changes per package
- **Changed Files**: Files section in report that shows detailed coverage changes per file

### Filtering Operations

- **Package Filtering**: Process of determining which packages appear in "Impacted Packages" based on threshold criteria
- **File Filtering**: Process of determining which files appear in "Changed files" based on exclusion threshold
- **Threshold-based Exclusion**: Filtering mechanism that uses percentage thresholds to control report content

