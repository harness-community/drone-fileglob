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

func setupRelativeFilesAndFolders() {
	fatalIf(os.MkdirAll("abc/def", 0755))
	fatalIf(os.WriteFile("abc/def/one.txt", []byte{}, 0644))
	fatalIf(os.WriteFile("abc/def/one.yml", []byte{}, 0644))
	fatalIf(os.WriteFile("abc/def/one.xml", []byte{}, 0644))
	fatalIf(os.WriteFile("abc/def/two.txt", []byte{}, 0644))
	fatalIf(os.WriteFile("abc/one.txt", []byte{}, 0644))
	fatalIf(os.WriteFile("abc/one.yml", []byte{}, 0644))
	fatalIf(os.WriteFile("abc/two.txt", []byte{}, 0644))

	fatalIf(os.WriteFile("a.xyz", []byte{}, 0644))
	fatalIf(os.WriteFile("b.xyz", []byte{}, 0644))
	fatalIf(os.WriteFile("a1.xyz", []byte{}, 0644))
	fatalIf(os.WriteFile("b1.xyz", []byte{}, 0644))

	fatalIf(os.MkdirAll("abc/test/harness/community", 0755))
	fatalIf(os.WriteFile("abc/test/harness/community/main.go", []byte{}, 0644))
	fatalIf(os.WriteFile("abc/test/harness/community/go.mod", []byte{}, 0644))
	fatalIf(os.WriteFile("abc/test/harness/community/go.sum", []byte{}, 0644))
}

func cleanupRelativeFilesAndFolders() {
	os.RemoveAll("abc")
	os.Remove("a.xyz")
	os.Remove("b.xyz")
	os.Remove("a1.xyz")
	os.Remove("b1.xyz")
}

func setupFilesAndFolders() string {
	tempDir, err := os.MkdirTemp("", "findfiles")
	fatalIf(err)

	fatalIf(os.MkdirAll(filepath.Join(tempDir, "abc/def"), 0755))
	fatalIf(os.WriteFile(filepath.Join(tempDir, "abc/def/one.txt"), []byte{}, 0644))
	fatalIf(os.WriteFile(filepath.Join(tempDir, "abc/def/one.yml"), []byte{}, 0644))
	fatalIf(os.WriteFile(filepath.Join(tempDir, "abc/def/one.xml"), []byte{}, 0644))
	fatalIf(os.WriteFile(filepath.Join(tempDir, "abc/def/two.txt"), []byte{}, 0644))
	fatalIf(os.WriteFile(filepath.Join(tempDir, "abc/one.txt"), []byte{}, 0644))
	fatalIf(os.WriteFile(filepath.Join(tempDir, "abc/one.yml"), []byte{}, 0644))
	fatalIf(os.WriteFile(filepath.Join(tempDir, "abc/two.txt"), []byte{}, 0644))
	fatalIf(os.WriteFile(filepath.Join(tempDir, "a.xyz"), []byte{}, 0644))
	fatalIf(os.WriteFile(filepath.Join(tempDir, "b.xyz"), []byte{}, 0644))
	fatalIf(os.WriteFile(filepath.Join(tempDir, "a1.xyz"), []byte{}, 0644))
	fatalIf(os.WriteFile(filepath.Join(tempDir, "b1.xyz"), []byte{}, 0644))

	fatalIf(os.MkdirAll(filepath.Join(tempDir, "abc/test/harness/community"), 0755))
	fatalIf(os.WriteFile(filepath.Join(tempDir, "abc/test/harness/community/main.go"), []byte{}, 0644))
	fatalIf(os.WriteFile(filepath.Join(tempDir, "abc/test/harness/community/go.mod"), []byte{}, 0644))
	fatalIf(os.WriteFile(filepath.Join(tempDir, "abc/test/harness/community/go.sum"), []byte{}, 0644))

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

// --
// ABSOLUTE PATTERN & PATH

func Test_Exec_Absolute_NoExcludes(t *testing.T) {
	tempDir := setupFilesAndFolders()
	defer os.RemoveAll(tempDir)

	args := Args{
		Filter:    "/**/*.txt",
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
	assert.Contains(t, paths, filepath.Join(tempDir, "abc/def/one.txt"))
	assert.Contains(t, paths, filepath.Join(tempDir, "abc/def/two.txt"))
	assert.Contains(t, paths, filepath.Join(tempDir, "abc/one.txt"))
	assert.Contains(t, paths, filepath.Join(tempDir, "abc/two.txt"))
}

func Test_Exec_Absolute_ExcludeDir(t *testing.T) {
	tempDir := setupFilesAndFolders()
	defer os.RemoveAll(tempDir)

	args := Args{
		Filter:    "/**/*.txt",
		Excludes:  "/**/def/*",
		TargetDir: tempDir,
	}

	files, err := applyFilter(NoopLogger(), args)
	assert.NoError(t, err)
	assert.Len(t, files, 2)

	var paths []string
	for _, file := range files {
		paths = append(paths, file.Path)
	}
	assert.Contains(t, paths, filepath.Join(tempDir, "abc/one.txt"))
	assert.Contains(t, paths, filepath.Join(tempDir, "abc/two.txt"))
}

func Test_Exec_Absolute_ExcludeFileExtension(t *testing.T) {
	tempDir := setupFilesAndFolders()
	defer os.RemoveAll(tempDir)

	args := Args{
		Filter:    "/**/def/*",
		Excludes:  "/**/*.txt",
		TargetDir: tempDir,
	}

	files, err := applyFilter(NoopLogger(), args)
	assert.NoError(t, err)
	assert.Len(t, files, 2)

	var paths []string
	for _, file := range files {
		paths = append(paths, file.Path)
	}
	assert.Contains(t, paths, filepath.Join(tempDir, "abc/def/one.yml"))
	assert.Contains(t, paths, filepath.Join(tempDir, "abc/def/one.xml"))
}

func Test_Exec_AbsoluteSingleCharacter(t *testing.T) {
	tempDir := setupFilesAndFolders()
	defer os.RemoveAll(tempDir)

	args := Args{
		Filter:    "/**/?.xyz",
		TargetDir: tempDir,
	}

	files, err := applyFilter(NoopLogger(), args)
	assert.NoError(t, err)
	assert.Len(t, files, 2)

	var paths []string
	for _, file := range files {
		paths = append(paths, file.Path)
	}
	assert.Contains(t, paths, filepath.Join(tempDir, "a.xyz"))
	assert.Contains(t, paths, filepath.Join(tempDir, "b.xyz"))
}

func Test_Exec_AbsoluteDirectoryWildcards(t *testing.T) {
	tempDir := setupFilesAndFolders()
	defer os.RemoveAll(tempDir)

	args := Args{
		Filter:    "/**/harness/**",
		TargetDir: tempDir,
	}

	files, err := applyFilter(NoopLogger(), args)
	assert.NoError(t, err)
	assert.Len(t, files, 4)

	var paths []string
	for _, file := range files {
		paths = append(paths, file.Path)
	}
	assert.Contains(t, paths, filepath.Join(tempDir, "abc/test/harness/community"))
	assert.Contains(t, paths, filepath.Join(tempDir, "abc/test/harness/community/main.go"))
	assert.Contains(t, paths, filepath.Join(tempDir, "abc/test/harness/community/go.mod"))
	assert.Contains(t, paths, filepath.Join(tempDir, "abc/test/harness/community/go.sum"))
}

// --
// RELATIVE PATTERN & PATH

func Test_Exec_RelativeSingleCharacter(t *testing.T) {
	setupRelativeFilesAndFolders()
	defer cleanupRelativeFilesAndFolders()

	args := Args{
		Filter: "?.xyz",
	}

	files, err := applyFilter(NoopLogger(), args)
	assert.NoError(t, err)
	assert.Len(t, files, 2)

	var paths []string
	for _, file := range files {
		paths = append(paths, file.Path)
	}
	assert.Contains(t, paths, "a.xyz")
	assert.Contains(t, paths, "b.xyz")
}

func Test_Exec_RelativeDirectoryWildcards(t *testing.T) {
	setupRelativeFilesAndFolders()
	defer cleanupRelativeFilesAndFolders()

	args := Args{
		Filter: "**/harness/**",
	}

	files, err := applyFilter(NoopLogger(), args)
	assert.NoError(t, err)
	assert.Len(t, files, 4)

	var paths []string
	for _, file := range files {
		paths = append(paths, file.Path)
	}
	assert.Contains(t, paths, "abc/test/harness/community")
	assert.Contains(t, paths, "abc/test/harness/community/main.go")
	assert.Contains(t, paths, "abc/test/harness/community/go.mod")
	assert.Contains(t, paths, "abc/test/harness/community/go.sum")
}

func Test_Exec_Relative_ExcludeFileExtension(t *testing.T) {
	setupRelativeFilesAndFolders()
	defer cleanupRelativeFilesAndFolders()

	args := Args{
		Filter:   "**/def/*",
		Excludes: "**/*.txt",
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

func Test_Exec_Relative_ExcludeDir(t *testing.T) {
	setupRelativeFilesAndFolders()
	defer cleanupRelativeFilesAndFolders()

	args := Args{
		Filter:   "**/*.txt",
		Excludes: "**/def/*",
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

func Test_Exec_Relative_NoExcludes(t *testing.T) {
	setupRelativeFilesAndFolders()
	defer cleanupRelativeFilesAndFolders()

	args := Args{
		Filter:   "**/*.txt",
		Excludes: "",
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

func NoopLogger() *logrus.Entry {
	log := logrus.New()
	log.SetOutput(io.Discard)

	return logrus.NewEntry(log)
}
