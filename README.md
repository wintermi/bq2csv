# BigQuery 2 CSV
[![Go Workflow Status](https://github.com/wintermi/bq2csv/workflows/Go/badge.svg)](https://github.com/wintermi/bq2csv/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/wintermi/bq2csv)](https://goreportcard.com/report/github.com/wintermi/bq2csv)
[![license](https://img.shields.io/github/license/wintermi/bq2csv.svg)](https://github.com/wintermi/bq2csv/blob/main/LICENSE)
[![GitHub release (latest by date including pre-releases)](https://img.shields.io/github/v/release/wintermi/bq2csv?include_prereleases)](https://github.com/wintermi/bq2csv/releases)


## Description
A command line application designed to provide a simple method to execute a BigQuery SQL script from "stdin", outputting all results to "stdout" in CSV format.  A detailed log is output to the console "stderr" providing you with the available execution statistics.

```
USAGE:
    bq2csv -p PROJECT_ID -d DATASET

ARGS:
  -c	Disable Query Cache
  -d string
    	BigQuery Dataset  (Required)
  -dr
    	Dry Run
  -f string
    	Field Delimter (default ",")
  -l string
    	BigQuery Data Processing Location
  -p string
    	Google Cloud Project ID  (Required)
  -v	Output Verbose Detail
```

## Example
```
echo "SELECT 1" | bq2csv -p PROJECT_ID -d DATASET 1> results.csv
```


## License
**bq2csv** is released under the [Apache License 2.0](https://github.com/wintermi/bq2csv/blob/main/LICENSE) unless explicitly mentioned in the file header.
