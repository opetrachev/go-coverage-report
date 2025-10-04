We need to change the behavior of the `go-coverage-report` tool. We will add a new command line parameter `-json`. When this parameter is specified, the tool should print the report in json format. See details below.

## Input data

Input data is in the `input-data` folder. These are just input files
for the tool. The tool processes the files and builds a report.

 * changed-files.json
 * coverage-1.out
 * coverage-2.out

The command is run with the following parameters:

`go run ./cmd/go-coverage-report -root github.com/fgrosse/go-coverage-report/ -trim github.com/fgrosse/go-coverage-report/ ./spec/2025-10-04\ add\ json\ output/input-data/coverage-1.out ./spec/2025-10-04\ add\ json\ output/input-data/coverage-2.out ./spec/2025-10-04\ add\ json\ output/input-data/changed-files.json`

## Current behavior

Currently when the tool is run, it prints a coverage report in 
the Markdown format. Examples are in the `old-behavior/report.md` file.

The report consists of 3 parts:

1. Coverage changes by package.
2. Details:
    2.1. Coverage info by file (only files with code).
    2.2. Changed unit tests files.

## New behavior

When the tool is run, and an additional `-json` parameter is passed, like this:

`go run ./cmd/go-coverage-report -json -root github.com/fgrosse/go-coverage-report/ -trim github.com/fgrosse/go-coverage-report/ ./spec/2025-10-04\ add\ json\ output/input-data/coverage-1.out ./spec/2025-10-04\ add\ json\ output/input-data/coverage-2.out ./spec/2025-10-04\ add\ json\ output/input-data/changed-files.json`

the tool should output a report with the same data, but in the form 
of json, like in the `new-behavior/report.json`. The report should be,
as before, printed to stdout.

In the provided `new-behavior/report.json` you can see an example of
the desired report, representing the same data as in the initial `old-behavior/report.md`.

Note that the json report does not include the `robot` column and any headers or text comments. However, it still includes all package names,
file names and numbers.

## What do you need to do

Please implement the new behavior (the ability to get the report in the json format), as described here.

