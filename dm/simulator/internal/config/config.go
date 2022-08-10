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

// Package config is the configuration definitions used by the simulator.
package config

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"reflect"

	"github.com/pingcap/errors"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v2"
)

// CLIConfig is the configuration struct for command-line-interface options.
type CLIConfig struct {
	IsHelp     bool
	ConfigFile string
}

// NewCLIConfig generates a new CLI config for being parsed.
func NewCLIConfig() *CLIConfig {
	theConf := new(CLIConfig)
	flag.StringVarP(&(theConf.ConfigFile), "config-file", "c", "", "config YAML file")
	flag.BoolVarP(&(theConf.IsHelp), "help", "h", false, "print help message")
	return theConf
}

// Config is the core configurations used by the simulator.
type Config struct {
	DataSources []*DataSourceConfig `yaml:"data_sources"`
	Workloads   []*WorkloadConfig   `yaml:"workloads"`
}

// NewConfigFromFile generates a new config object from a configuration file.
// The config file is in YAML format.
func NewConfigFromFile(configFile string) (*Config, error) {
	f, err := os.Open(configFile)
	if err != nil {
		return nil, errors.Annotate(err, "open config file error")
	}
	theConfig := new(Config)
	dec := yaml.NewDecoder(f)
	err = dec.Decode(theConfig)
	if err != nil {
		return nil, errors.Annotate(err, "decode YAML error")
	}
	return theConfig, nil
}

// DataSourceConfig is the sub config for describing a DB data source.
type DataSourceConfig struct {
	DataSourceID string         `yaml:"id"`
	Host         string         `yaml:"host"`
	Port         int            `yaml:"port"`
	UserName     string         `yaml:"user"`
	Password     string         `yaml:"password"`
	Tables       []*TableConfig `yaml:"tables"`
}

// TableConfig is the sub config for describing a simulating table in the data source.
type TableConfig struct {
	TableID              string              `yaml:"id"`
	DatabaseName         string              `yaml:"db"`
	TableName            string              `yaml:"table"`
	Columns              []*ColumnDefinition `yaml:"columns"`
	UniqueKeyColumnNames []string            `yaml:"unique_keys"`
}

// SortedClone clones a table config with the columns sorted.
func (cfg *TableConfig) SortedClone() *TableConfig {
	return &TableConfig{
		TableID:              cfg.TableID,
		DatabaseName:         cfg.DatabaseName,
		TableName:            cfg.TableName,
		Columns:              CloneSortedColumnDefinitions(cfg.Columns),
		UniqueKeyColumnNames: append([]string{}, cfg.UniqueKeyColumnNames...),
	}
}

// CreateUpstreamTable creates tables in the upstream database.
func (cfg *TableConfig) CreateUpstreamTable(ctx context.Context, db *sql.DB) error {
	sql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", cfg.DatabaseName)
	_, err := db.ExecContext(ctx, sql)
	if err != nil {
		return err
	}
	columnStr := ""
	for _, column := range cfg.Columns {
		columnStr = fmt.Sprintf("%s%s %s", columnStr, column.ColumnName, column.DataType)
		if column.Length != 0 {
			columnStr = fmt.Sprintf("%s(%d)", columnStr, column.Length)
		}
		columnStr = fmt.Sprintf("%s, ", columnStr)
	}
	ukStr := ""
	for i, ukName := range cfg.UniqueKeyColumnNames {
		if i == len(cfg.UniqueKeyColumnNames)-1 {
			ukStr = fmt.Sprintf("%s%s", ukStr, ukName)
		} else {
			ukStr = fmt.Sprintf("%s%s, ", ukStr, ukName)
		}
	}
	columnStr = fmt.Sprintf("%sUNIQUE KEY (%s)", columnStr, ukStr)
	sql = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (%s);", cfg.DatabaseName, cfg.TableName, columnStr)
	_, err = db.ExecContext(ctx, sql)
	if err != nil {
		return err
	}
	return nil
}

// DropUpstreamTable drops tables in the upstream database.
func (cfg *TableConfig) DropUpstreamTable(ctx context.Context, db *sql.DB) error {
	sql := fmt.Sprintf("DROP TABLE IF EXISTS %s.%s;", cfg.DatabaseName, cfg.TableName)
	_, err := db.ExecContext(ctx, sql)
	if err != nil {
		return err
	}
	return nil
}

// IsValid checks whether the table format is valid.
func (cfg *TableConfig) IsValid() (bool, error) {
	if len(cfg.Columns) == 0 {
		return false, errors.New("no columns assigned")
	}
	columnNameMap := make(map[string]bool)
	for _, column := range cfg.Columns {
		if ok := columnNameMap[column.ColumnName]; ok {
			return false, errors.New(fmt.Sprintf("duplicate column names: %s", column.ColumnName))
		}
		columnNameMap[column.ColumnName] = true
		switch column.DataType {
		case "varchar":
			if column.Length < 100 {
				return false, errors.New(fmt.Sprintf("column %s has string type with length %d, but at least length 100 is required", column.ColumnName, column.Length))
			}
		case "text":
		case "int":
		case "timestamp", "datetime":
		case "float":
		default:
			return false, errors.New(fmt.Sprintf("unknown data type: %s", column.DataType))
		}
	}
	if len(cfg.UniqueKeyColumnNames) == 0 {
		return false, errors.New("no unique keys assigned")
	}
	for _, ukName := range cfg.UniqueKeyColumnNames {
		hasCorrespondColumn := false
		for _, column := range cfg.Columns {
			if ukName == column.ColumnName {
				hasCorrespondColumn = true
				break
			}
		}
		if !hasCorrespondColumn {
			return false, errors.New(fmt.Sprintf("no correspond column for unique key: %s", ukName))
		}
	}
	return true, nil
}

// IsDeepEqual compares two tables configs to see whether all the values are equal.
func (cfg *TableConfig) IsDeepEqual(cfgB *TableConfig) bool {
	if cfg == nil || cfgB == nil {
		return false
	}
	if cfg.TableID != cfgB.TableID ||
		cfg.DatabaseName != cfgB.DatabaseName ||
		cfg.TableName != cfgB.TableName {
		return false
	}
	if !AreColDefinitionsEqual(cfg.Columns, cfgB.Columns) {
		return false
	}
	// begin to check the unique key names
	if !reflect.DeepEqual(cfg.UniqueKeyColumnNames, cfgB.UniqueKeyColumnNames) {
		return false
	}
	return true
}

// ColumnDefinition is the sub config for describing a column in a simulating table.
type ColumnDefinition struct {
	ColumnName string `yaml:"name"`
	DataType   string `yaml:"type"`
	Length     int    `yaml:"length"`
}

// WorkloadConfig is the configuration to describe the attributes of a workload.
type WorkloadConfig struct {
	DataSources  []string `yaml:"data_sources"`
	WorkloadCode string   `yaml:"dsl_code"`
}
