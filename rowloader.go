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
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"cloud.google.com/go/bigquery"
)

// RowLoader implements bigquery.ValueLoader
type RowLoader struct {
	// The BigQuery Schema for the row
	Schema bigquery.Schema

	// The converted row
	Row []string
}

var _ bigquery.ValueLoader = &RowLoader{}

//---------------------------------------------------------------------------------------

// Load implements bigquery.ValueLoader.
func (r *RowLoader) Load(row []bigquery.Value, schema bigquery.Schema) error {
	r.Row = make([]string, len(row))
	r.Schema = schema

	for i, val := range row {
		switch val := val.(type) {
		case string:
			r.Row[i] = val
		case int64:
			r.Row[i] = strconv.FormatInt(val, 10)
		case *big.Rat:
			switch schema[i].Type {
			case bigquery.NumericFieldType:
				r.Row[i] = strings.TrimRight(strings.TrimRight(bigquery.NumericString(val), "0"), ".")
			case bigquery.BigNumericFieldType:
				r.Row[i] = strings.TrimRight(strings.TrimRight(bigquery.BigNumericString(val), "0"), ".")
			default:
				r.Row[i] = fmt.Sprint(val)
			}
		default:
			r.Row[i] = fmt.Sprint(val)
		}
	}

	return nil
}
