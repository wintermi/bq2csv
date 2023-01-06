// Copyright 2022-2023, Matthew Winter
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
)

var logger zerolog.Logger
var applicationText = "%s 0.1.1%s"
var copyrightText = "Copyright 2022-2023, Matthew Winter\n"
var indent = "..."

var helpText = `
A command line application designed to provide a simple method to execute a
BigQuery SQL script from "stdin", outputting all results to "stdout" in CSV
format.  A detailed log is output to the console "stderr" providing you with
the available execution statistics.

Use --help for more details.


USAGE:
    bq2csv -p PROJECT_ID -d DATASET

ARGS:
`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, applicationText, filepath.Base(os.Args[0]), "\n")
		fmt.Fprint(os.Stderr, copyrightText)
		fmt.Fprint(os.Stderr, helpText)
		flag.PrintDefaults()
	}

	// Define the Long CLI flag names
	var targetProject = flag.String("p", "", "Google Cloud Project ID  (Required)")
	var targetDataset = flag.String("d", "", "BigQuery Dataset  (Required)")
	var fieldDelimiter = flag.String("f", ",", "Field Delimter")
	var processingLocation = flag.String("l", "", "BigQuery Data Processing Location")
	var disableQueryCache = flag.Bool("c", false, "Disable Query Cache")
	var dryRun = flag.Bool("dr", false, "Dry Run")
	var verbose = flag.Bool("v", false, "Output Verbose Detail")

	// Parse the flags
	flag.Parse()

	// Validate the Required Flags
	if *targetProject == "" || *targetDataset == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Validate that the Field Delimiter is 1 character
	if len(*fieldDelimiter) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	// Setup Zero Log for Consolo Output
	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	logger = zerolog.New(output).With().Timestamp().Logger()
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000"
	zerolog.DurationFieldUnit = time.Millisecond
	zerolog.DurationFieldInteger = true
	if *verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Output Header
	logger.Info().Msgf(applicationText, filepath.Base(os.Args[0]), "")
	logger.Info().Msg("Arguments")
	logger.Info().Str("Project ID", *targetProject).Msg(indent)
	logger.Info().Str("Dataset", *targetDataset).Msg(indent)
	logger.Info().Str("Field Delimiter", *fieldDelimiter).Msg(indent)
	logger.Info().Str("Processing Location", *processingLocation).Msg(indent)
	logger.Info().Bool("Disable Query Cache", *disableQueryCache).Msg(indent)
	logger.Info().Bool("Dry Run", *dryRun).Msg(indent)
	logger.Info().Msg("Begin")

	// Load the BigQuery SQL into memory ready for execution
	var query Query
	err := query.ReadStdIn()
	if err != nil {
		logger.Error().Err(err).Msg("Check STDIN, No SQL Found")
		os.Exit(1)
	}

	// Check that something exists in SQL
	if len(query.SQL) == 0 {
		logger.Error().Msg("Check STDIN, No SQL Found")
		os.Exit(1)
	}
	logger.Info().Int("SQL Length", len(query.SQL)).Msg("Reading SQL Complete")

	// Execute the SQL outputting results to the StdOut
	err = query.ExecuteQueries(*targetProject, *targetDataset, *processingLocation, *disableQueryCache, *dryRun, *fieldDelimiter)
	if err != nil {
		logger.Error().Err(err).Msg("SQL Execution Failed")
		os.Exit(1)
	}
	logger.Info().Msg("End")
}
