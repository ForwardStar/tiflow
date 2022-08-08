// Copyright 2022 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package core

import (
	"github.com/pingcap/errors"
)

var (
	// ErrUnsupportedColumnType means the column type is unsupported for retrieving data from the DB.
	// It is used when loading the existing DB data into the MCP.
	ErrUnsupportedColumnType error = errors.New("unsupported column type")

	// ErrTableConfigNotFound means the table configuration is not found.
	ErrTableConfigNotFound error = errors.New("table config not found")
)
