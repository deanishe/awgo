// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func inTempDir(fun func(dir string)) error {
	curdir, err := os.Getwd()
	if err != nil {
		return err
	}

	dir, err := ioutil.TempDir("", "awgo-util-")
	if err != nil {
		return err
	}
	// TempDir() returns a symlink on my macOS :(
	if dir, err = filepath.EvalSymlinks(dir); err != nil {
		return err
	}

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			panic(err)
		}
	}()

	// Change to temporary directory
	if err := os.Chdir(dir); err != nil {
		return err
	}

	// Change back after we're done
	defer func() {
		if err := os.Chdir(curdir); err != nil {
			panic(err)
		}
	}()

	fun(dir)

	return nil
}

func TestMustExist(t *testing.T) {
	err := inTempDir(func(dir string) {
		name := "testdir"

		// Create directory
		s := MustExist(name)
		assert.Equal(t, name, s, "unexpected dirname")

		_, err := os.Stat(s)
		assert.Nil(t, err, "stat dir failed")

		// Check path is as expected
		p := filepath.Join(dir, name)
		p2, err := filepath.Abs(s)
		require.Nil(t, err, "filepath.Abs failed")
		assert.Equal(t, p, p2, "unexpected path")
	})

	require.Nil(t, err, "inTempDir failed")
}

func TestPathExists(t *testing.T) {
	t.Parallel()

	err := inTempDir(func(dir string) {
		name := "existingdir"
		path := filepath.Join(dir, name)
		badName := "nodir"
		badPath := filepath.Join(dir, badName)

		require.Nil(t, os.MkdirAll(name, 0700), "MkdirAll failed")

		data := []struct {
			p string
			x bool
		}{
			{dir, true},
			{name, true},
			{path, true},
			{badName, false},
			{badPath, false},
		}

		for _, td := range data {
			assert.Equal(t, td.x, PathExists(td.p), "unexpected result")
		}
	})

	require.Nil(t, err, "inTempDir failed")
}

func TestClearDirectory(t *testing.T) {
	err := inTempDir(func(dir string) {
		names := []string{"./root/one", "./root/two", "./root/three"}
		for _, s := range names {
			require.Nil(t, os.MkdirAll(s, 0700), "MkdirAll failed")
		}

		for _, s := range names {
			_, err := os.Stat(s)
			assert.Nil(t, err, "stat failed")
		}
		assert.Nil(t, ClearDirectory("./root"), "ClearDirectory failed")

		for _, s := range names {
			s := s
			t.Run(s, func(t *testing.T) {
				_, err := os.Stat(s)
				assert.True(t, os.IsNotExist(err), "file exists")
			})
		}
	})
	require.Nil(t, err, "inTempDir failed")
}

func TestWriteFile(t *testing.T) {
	err := inTempDir(func(dir string) {
		var (
			name    = "test.txt"
			content = []byte(`test`)
		)

		require.False(t, PathExists(name), "path already exists")
		require.Nil(t, WriteFile(name, content, 0600), "WriteFile failed")
		require.True(t, PathExists(name), "path does not exist")

		data, err := ioutil.ReadFile(name)
		require.Nil(t, err, "read file failed")

		assert.Equal(t, content, data, "unexpected file content")

		fi, err := os.Stat(name)
		require.Nil(t, err, "stat file failed")
		assert.Equal(t, os.FileMode(0600), fi.Mode(), "unexpected file mode")

		infos, err := ioutil.ReadDir(".")
		require.Nil(t, err, "ReadDir failed")
		assert.Equal(t, 1, len(infos), "unexpected no. of files")
	})
	require.Nil(t, err, "inTempDir failed")
}
