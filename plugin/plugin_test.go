// Copyright 2020 the Drone Authors. All rights reserved.
// Use of this source code is governed by the Blue Oak Model License
// that can be found in the LICENSE file.

package plugin

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func fatalIf(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func setupFilesAndFolders() string {
	tempDir, err := os.MkdirTemp("", "fileglob")
	fatalIf(err)

	fatalIf(os.MkdirAll(filepath.Join(tempDir, "abc/def"), 0755))
	fatalIf(os.WriteFile(filepath.Join(tempDir, "abc/def/one.txt"), []byte{}, 0644))
	fatalIf(os.WriteFile(filepath.Join(tempDir, "abc/def/one.yml"), []byte{}, 0644))
	fatalIf(os.WriteFile(filepath.Join(tempDir, "abc/def/one.xml"), []byte{}, 0644))
	fatalIf(os.WriteFile(filepath.Join(tempDir, "abc/def/two.txt"), []byte{}, 0644))
	fatalIf(os.WriteFile(filepath.Join(tempDir, "abc/one.txt"), []byte{}, 0644))
	fatalIf(os.WriteFile(filepath.Join(tempDir, "abc/one.yml"), []byte{}, 0644))
	fatalIf(os.WriteFile(filepath.Join(tempDir, "abc/two.txt"), []byte{}, 0644))

	return tempDir
}

func Test_validateArg_SunnyDay(t *testing.T) {
	err := validateArgs(Args{
		Filter: "**/*.txt",
	})
	assert.NoError(t, err)
}

func Test_validateArg_MissingFilter(t *testing.T) {
	err := validateArgs(Args{})

	assert.EqualError(t, err, "filter glob is empty")
}

func Test_fileInfo_FileNotExist(t *testing.T) {
	_, err := getFileInfo("file-not-exist.txt")

	assert.Error(t, err)
}

func Test_fileInfo_Directory(t *testing.T) {
	tempDir := setupFilesAndFolders()
	defer os.RemoveAll(tempDir)

	expectedTime := time.Now().Format(time.RFC3339)

	path := filepath.Join(tempDir, "abc/def")
	fi, err := getFileInfo(path)
	fatalIf(err)

	assert.Equal(t, "def", fi.Name)
	assert.Equal(t, path, fi.Path)
	assert.True(t, fi.IsDirectory)
	assert.NotEqualValues(t, int64(0), fi.Length)
	assert.Equal(t, expectedTime, fi.LastModified)
}

func Test_fileInfo_File(t *testing.T) {
	tempDir := setupFilesAndFolders()
	defer os.RemoveAll(tempDir)

	expectedTime := time.Now().Format(time.RFC3339)

	path := filepath.Join(tempDir, "abc/one.txt")
	fi, err := getFileInfo(path)
	fatalIf(err)

	assert.Equal(t, "one.txt", fi.Name)
	assert.Equal(t, path, fi.Path)
	assert.False(t, fi.IsDirectory)
	assert.Equal(t, int64(0), fi.Length)
	assert.Equal(t, expectedTime, fi.LastModified)
}

func Test_Exec_NoExcludes(t *testing.T) {
	tempDir := setupFilesAndFolders()
	defer os.RemoveAll(tempDir)

	curdir, err := os.Getwd()
	fatalIf(err)
	fatalIf(os.Chdir(tempDir))
	defer os.Chdir(curdir)

	args := Args{
		Filter:    "**/*.txt",
		Excludes:  "",
		TargetDir: tempDir,
	}

	files, err := applyFilter(NoopLogger(), args)
	assert.NoError(t, err)
	assert.Len(t, files, 4)

	var paths []string
	for _, file := range files {
		paths = append(paths, file.Path)
	}
	assert.Contains(t, paths, "abc/def/one.txt")
	assert.Contains(t, paths, "abc/def/two.txt")
	assert.Contains(t, paths, "abc/one.txt")
	assert.Contains(t, paths, "abc/two.txt")
}

func Test_Exec_ExcludeDir(t *testing.T) {
	tempDir := setupFilesAndFolders()
	defer os.RemoveAll(tempDir)

	curdir, err := os.Getwd()
	fatalIf(err)
	fatalIf(os.Chdir(tempDir))
	defer os.Chdir(curdir)

	args := Args{
		Filter:    "**/*.txt",
		Excludes:  "**/def/*",
		TargetDir: tempDir,
	}

	files, err := applyFilter(NoopLogger(), args)
	assert.NoError(t, err)
	assert.Len(t, files, 2)

	var paths []string
	for _, file := range files {
		paths = append(paths, file.Path)
	}
	assert.Contains(t, paths, "abc/one.txt")
	assert.Contains(t, paths, "abc/two.txt")
}

func Test_Exec_ExcludeFileExtension(t *testing.T) {
	tempDir := setupFilesAndFolders()
	defer os.RemoveAll(tempDir)

	curdir, err := os.Getwd()
	fatalIf(err)
	fatalIf(os.Chdir(tempDir))
	defer os.Chdir(curdir)

	args := Args{
		Filter:    "**/def/*",
		Excludes:  "**/*.txt",
		TargetDir: tempDir,
	}

	files, err := applyFilter(NoopLogger(), args)
	assert.NoError(t, err)
	assert.Len(t, files, 2)

	var paths []string
	for _, file := range files {
		paths = append(paths, file.Path)
	}
	assert.Contains(t, paths, "abc/def/one.yml")
	assert.Contains(t, paths, "abc/def/one.xml")
}

func NoopLogger() *logrus.Entry {
	log := logrus.New()
	log.SetOutput(io.Discard)

	return logrus.NewEntry(log)
}
