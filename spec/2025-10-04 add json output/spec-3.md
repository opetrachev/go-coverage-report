## AS IS

Currently, when a `-format json-combined` is used, it creates the
`coverage_by_file` section. This section lists all files in a plain list, exmple:

```
    "coverage_by_file": [
        {
            "name": "cmd/go-coverage-report/changes/changed_files.go",
            "coverage": 80,
            "change": 80,
            "total": 10,
            "covered": 8,
            "covered_change": 8,
            "missed": 2,
            "missed_change": -8
        },
        {
            "name": "cmd/go-coverage-report/changes/changed_unittests.go",
            "coverage": 80,
            "change": 80,
            "total": 10,
            "covered": 8,
            "covered_change": 8,
            "missed": 2,
            "missed_change": -8
        },
        {
            "name": "cmd/go-coverage-report/reporting/report.go",
            "coverage": 0,
            "change": 0,
            "total": 0,
            "covered": 0,
            "covered_change": 0,
            "missed": 0,
            "missed_change": 0
        }
    ]
```

## TO BE

Files in the `coverage_by_file` should be grouped by package, like this:

```
    "coverage_by_file": [
        {
            "package": "cmd/go-coverage-report/changes",
            "files": [
                {
                    "name": "changed_files.go",
                    "coverage": 80,
                    "change": 80,
                    "total": 10,
                    "covered": 8,
                    "covered_change": 8,
                    "missed": 2,
                    "missed_change": -8
                },
                {
                    "name": "changed_unittests.go",
                    "coverage": 80,
                    "change": 80,
                    "total": 10,
                    "covered": 8,
                    "covered_change": 8,
                    "missed": 2,
                    "missed_change": -8
                },
            ],  
        },
        {
            "package": "cmd/go-coverage-report/reporting",
            "files": [
                {
                    "name": "report.go",
                    "coverage": 0,
                    "change": 0,
                    "total": 0,
                    "covered": 0,
                    "covered_change": 0,
                    "missed": 0,
                    "missed_change": 0
                }
            ]
        }
    ]
```

## Your task

Implement the new behavior.
