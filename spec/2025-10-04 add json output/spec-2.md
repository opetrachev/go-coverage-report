## Current behavior

When `-format json` command line option is specified, the tool produces report in the json format.

## Problem

Before the last change, the tool already was able to produce report in the json format with `-format json` parameter. The new 
report format replaced the old report format, and the old format
is unavailable now.

## Solution

Revert back to the old behavior, where `-format json` produces the original json report.

For the new report format, add anoa new flag value `-format json-combined`, which will produce report in the new format.

## What to do

Implement the fix.