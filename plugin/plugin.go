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
	"time"

	"github.com/georgeJobs/go-antpathmatcher"
	"github.com/sirupsen/logrus"
)

// Args provides plugin execution arguments.
type Args struct {
	Pipeline

	// Level defines the plugin log level.
	Level string `envconfig:"PLUGIN_LOG_LEVEL"`

	// Ant style pattern Glob pattern to search for files. (required)
	Filter string `envconfig:"PLUGIN_FILTER"`

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
	LastModified string `json:"lastModified"`
}

// Exec executes the plugin.
func Exec(ctx context.Context, args Args) error {
	if err := validateArgs((args)); err != nil {
		return err
	}

	logger := logrus.
		WithField("glob", args.Filter).
		WithField("excludes", args.Excludes).
		WithField("dir", args.TargetDir)
	logger.Infoln("searching files")

	files, err := applyFilter(logger, args)
	if err != nil {
		return err
	}

	jsonOutput, err := json.Marshal(files)
	if err != nil {
		return logError(logger, fmt.Sprintf("Error marshalling JSON: %v", err), err)
	}

	if err = writeEnvToFile("FILES_INFO", string(jsonOutput)); err != nil {
		return err
	}
	return nil
}

func applyFilter(logger *logrus.Entry, args Args) ([]FileInfo, error) {
	var files []FileInfo
	m := antpathmatcher.NewAntPathMatcher()

	if args.TargetDir == "" {
		args.TargetDir = "."
	}

	err := filepath.WalkDir(args.TargetDir, func(path string, d os.DirEntry, e error) error {

		if m.Match(args.Filter, path) {
			if m.Match(args.Excludes, path) {
				logger.Debugf("path %s match exclude criteria %s", path, args.Excludes)

			} else {
				file, err := getFileInfo(path)
				if err != nil {
					return logError(logger, fmt.Sprintf("error to get file info of path %s", path), err)
				}

				files = append(files, file)
			}
		}

		return nil
	})
	if err != nil {
		return []FileInfo{}, err
	}

	return files, nil
}

func getFileInfo(path string) (FileInfo, error) {

	// RETRIEVE DETAILS ABOUT THE PROVIDED PATH
	fi, err := os.Lstat(path)

	if err != nil {
		return FileInfo{}, err
	}
	return FileInfo{
		Name:         fi.Name(),
		Path:         path,
		IsDirectory:  fi.IsDir(),
		Length:       fi.Size(),
		LastModified: fi.ModTime().Format(time.RFC3339),
	}, nil
}

func logError(logger *logrus.Entry, message string, err error) error {
	logger.Error(message)
	return err
}

func validateArgs(args Args) error {
	if args.Filter == "" {
		return errors.New("filter is empty")
	}
	if os.Getenv("DRONE_OUTPUT") == "" {
		return errors.New("missing DRONE_OUTPUT environment variable")
	}
	return nil
}

// writeEnvToFile uses the Drone environment variable DRONE_OUTPUT to write the search result
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
