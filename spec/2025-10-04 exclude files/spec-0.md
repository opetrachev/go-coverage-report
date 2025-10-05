## Exclude files feature

I want to add a feature to exclude files from the report. I think it should be a new command line flag. 

To clarify the necessary details, you will ask me questions in chat in russian, and I will provide you with answers. You will save information, important for the implementation, here under the Discussion section, in english.

Ask me from time to time if we need anything else to discuss or suggest your questions, don't start the implementation until we discuss all the necessary implementation aspects.

## Discussion

### Implementation details discussed:

1. **Flag format**: `--exclude-file "path/to/file"` - reads exclusion patterns from a file
2. **Pattern support**: Glob patterns with paths (e.g., `github.com/user/project/*_test.go`)
3. **Scope**: Exclusions applied at Coverage struct level - files should not exist there at all
4. **Impact**: Excluded files should not affect package TotalStmt, coverage calculations, etc.
5. **Interaction with other flags**: 
   - `-root`: No interaction (relates to changed_files, not coverage.out)
   - `-trim`: No interaction (applied after coverage loading)
6. **Multiple exclusions**: Handled via file with multiple patterns (one per line?)

### Final implementation decisions:

1. **File format**: Like .gitignore format (one pattern per line, comments with #, empty lines ignored)
2. **Error handling**: Standard error handling - exit with error on file not found or invalid patterns
3. **Application scope**: 
   - Exclude files from both `Old.Files` and `New.Files`
   - Don't include excluded files in initial parsing rather than recalculating after exclusion
   - Excluded files should not appear in markdown or JSON reports (as if they never existed)
   - **IMPORTANT**: Also exclude files from `changed files` - files excluded from coverage should not appear in the changed files list in the report
4. **Timing**: Apply exclusions during coverage file loading (in `ParseCoverage`)
5. **Testing**: Add tests for happy path scenarios
6. **Architecture**: 
   - Pass exclude patterns as parameter to `ParseCoverage(filename string, excludePatterns []string) (*Coverage, error)`
   - Also update `ParseChangedFiles` to accept exclude patterns: `ParseChangedFiles(filename, prefix string, excludePatterns []string) ([]string, error)`