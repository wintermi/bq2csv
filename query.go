// Copyright 2022, Matthew Winter
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
	"context"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
)

type Query struct {
	SQL                  string
	Error                error
	QueryStartTime       time.Time
	QueryEndTime         time.Time
	FirstRowReturnedTime time.Time
	AllRowsReturnedTime  time.Time
	TotalRowsReturned    int64
}

//---------------------------------------------------------------------------------------

// Read the BigQuery SQL into memory from STDIN ready for execution
func (sql *Query) ReadStdIn() error {

	// Check if there is somethinig to read on STDIN, return error if not
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return fmt.Errorf("[ReadStdIn] No SQL Found")
	}

	// Read data from STDIN
	buf, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return fmt.Errorf("[ReadStdIn] Read All Failed: %w", err)
	}

	sql.SQL = string(buf)

	logger.Debug().Msg("Query Details")
	logger.Debug().Str("SQL", sql.SQL).Msg(indent)

	return nil
}

//---------------------------------------------------------------------------------------

// Execute the SQL in BigQuery
func (sql *Query) ExecuteQueries(project string, dataset string, location string, cache bool, dryRun bool) error {

	// Establish a BigQuery Client Connection
	logger.Info().Msg("Establishing a BigQuery Client Connection")
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, project)
	if err != nil {
		return fmt.Errorf("Failed Establishing a BigQuery Client Connection: %w", err)
	}
	defer client.Close()

	// BigQuery Client Configuration
	client.Location = location

	// Execute SQL
	if dryRun {
		sql.ExecuteDryRun(ctx, client, project, dataset, cache)
		sql.LogExecuteDryRun()
	} else {
		sql.ExecuteQuery(ctx, client, project, dataset, cache)
		sql.LogExecuteQuery()
	}

	// Raise an Error if query execution failed
	if sql.Error != nil {
		return fmt.Errorf("One or More Queries Failed")
	}

	return nil
}

//---------------------------------------------------------------------------------------

// Execute Query
func (sql *Query) ExecuteQuery(ctx context.Context, client *bigquery.Client, project string, dataset string, cache bool) {
	// Create and Configure Query
	q := client.Query(sql.SQL)
	q.DefaultProjectID = project
	q.DefaultDatasetID = dataset
	q.DisableQueryCache = !cache
	q.DryRun = false

	// Initiate the Query Job
	sql.QueryStartTime = time.Now()
	it, err := q.Read(ctx)
	sql.QueryEndTime = time.Now()
	if err != nil {
		sql.Error = err
		return
	}

	// Ready the CSV Writer
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	var row []bigquery.Value
	var rowCount int64
	for {
		err := it.Next(&row)
		if rowCount == 0 {
			sql.FirstRowReturnedTime = time.Now()
		}
		if err == iterator.Done {
			sql.AllRowsReturnedTime = time.Now()
			sql.TotalRowsReturned = rowCount
			break
		}
		if err != nil {
			sql.Error = err
			return
		}
		if err := w.Write(*bqToString(&row)); err != nil {
			sql.Error = fmt.Errorf("Failed Writing to the Output File")
			return
		}
		rowCount++
	}
}

//---------------------------------------------------------------------------------------

// Execute Dry Run Query
func (sql *Query) ExecuteDryRun(ctx context.Context, client *bigquery.Client, project string, dataset string, cache bool) {
	// Create and Configure Query
	q := client.Query(sql.SQL)
	q.DefaultProjectID = project
	q.DefaultDatasetID = dataset
	q.DisableQueryCache = !cache
	q.DryRun = true

	// Initiate the Query Job
	sql.QueryStartTime = time.Now()
	job, err := q.Run(ctx)
	if err != nil {
		sql.Error = err
		return
	}

	// Check the Last Status for Errors
	status := job.LastStatus()
	if err = status.Err(); err != nil {
		sql.Error = err
		return
	}
	sql.QueryEndTime = time.Now()
}

//---------------------------------------------------------------------------------------

// Output the Query Execution Statistics to the Log
func (sql *Query) LogExecuteQuery() {
	logger.Info().Msg("Query Execution")

	// Output Error Message if one exists, but nothing else
	if sql.Error != nil {
		logger.Error().Err(sql.Error).Msg(indent)
		return
	}

	logger.Info().Time("Query Execution Start", sql.QueryStartTime).Msg(indent)
	logger.Info().Time("Query Execution End", sql.QueryEndTime).Msg(indent)
	logger.Info().TimeDiff("Execution Time (ms)", sql.QueryEndTime, sql.QueryStartTime).Msg(indent)
	logger.Info().Time("First Row Returned", sql.FirstRowReturnedTime).Msg(indent)
	logger.Info().Time("All Rows Returned", sql.AllRowsReturnedTime).Msg(indent)
	logger.Info().TimeDiff("Return Time (ms)", sql.AllRowsReturnedTime, sql.QueryEndTime).Msg(indent)
	logger.Info().Int64("Total Rows Returned", sql.TotalRowsReturned).Msg(indent)
}

//---------------------------------------------------------------------------------------

// Output the Query Dry Run Statistics to the Log
func (sql *Query) LogExecuteDryRun() {
	logger.Info().Msg("Query Dry Run")

	// Output Error Message if one exists, but nothing else
	if sql.Error != nil {
		logger.Error().Err(sql.Error).Msg(indent)
		return
	}

	logger.Info().Time("Query Execution Start", sql.QueryStartTime).Msg(indent)
	logger.Info().Time("Query Execution End", sql.QueryEndTime).Msg(indent)
	logger.Info().TimeDiff("Execution Time (ms)", sql.QueryEndTime, sql.QueryStartTime).Msg(indent)
}

//---------------------------------------------------------------------------------------

// Convert BigQuery.Value Array to a String Array
func bqToString(row *[]bigquery.Value) *[]string {
	record := make([]string, len(*row))

	for i, val := range *row {
		record[i] = fmt.Sprint(val)
	}
	return &record
}
