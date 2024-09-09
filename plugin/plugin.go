// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

// Args provides plugin execution arguments.
type Args struct {
	Pipeline

	// Level defines the plugin log level.
	Level string `envconfig:"PLUGIN_LOG_LEVEL"`

	// Ant style pattern Glob pattern to search for files. (required)
	Filter string `envconfig:"PLUGIN_GLOB"`

	// Glob pattern to exclude files from the search (optional) (default: none)
	Excludes string `envconfig:"PLUGIN_EXCLUDES"`

	// Directory in which to perform the search. If not specified, the current directory is used. (optional)
	TargetDir string `envconfig:"PLUGIN_DIR"`
}

type FileInfo struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	IsDirectory  bool   `json:"isDirectory"`
	Length       int64  `json:"length"`
	LastModified int64  `json:"lastModified"`
}

// Exec executes the plugin.
func Exec(ctx context.Context, args Args) error {

	if err := validateArgs((args)); err != nil {
		return err
	}

	if args.TargetDir == "" {
		args.TargetDir = "."
	}

	logger := logrus.
		WithField("glob", args.Filter).
		WithField("excludes", args.Excludes).
		WithField("dir", args.TargetDir)
	logger.Infoln("searching files")

	var files []FileInfo

	err := filepath.Walk(args.TargetDir, func(path string, info os.FileInfo, e error) error {

		if e != nil {
			return e
		}

		ok, err := filepath.Match(args.Filter, path)
		if err != nil {
			return err
		}
		if ok {
			fmt.Printf("match %s\n", path)

			file := FileInfo{
				Name:         info.Name(),
				Path:         path,
				IsDirectory:  info.IsDir(),
				Length:       info.Size(),
				LastModified: info.ModTime().Unix(),
			}
			files = append(files, file)
		}

		return nil
	})
	if err != nil {
		logger.Fatal(err)
	}

	jsonOutput, err := json.Marshal(files)
	if err != nil {
		logger.Errorf("Error marshalling JSON: %v", err)
		return err
	}

	if err = writeEnvToFile("FILES_INFO", string(jsonOutput)); err != nil {
		return err
	}
	return nil
}

func validateArgs(args Args) error {
	if args.Filter == "" {
		return errors.New("filter glob is empty")
	}
	return nil
}

func writeEnvToFile(key, value string) error {
	outputFile, err := os.OpenFile(os.Getenv("DRONE_OUTPUT"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	defer outputFile.Close()

	_, err = fmt.Fprintf(outputFile, "%s=%s\n", key, value)
	if err != nil {
		return fmt.Errorf("failed to write to env: %w", err)
	}

	return nil
}
