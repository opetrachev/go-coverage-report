package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func ParseChangedFiles(filename, prefix string, excludePatterns []string) ([]string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var files []string
	err = json.Unmarshal(data, &files)
	if err != nil {
		return nil, err
	}

	for i, file := range files {
		files[i] = filepath.Join(prefix, file)
	}

	// Filter out excluded files
	if len(excludePatterns) > 0 {
		files = filterExcludedFiles(files, excludePatterns)
	}

	return files, nil
}

// filterExcludedFiles filters out files that match exclusion patterns
func filterExcludedFiles(files []string, excludePatterns []string) []string {
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
