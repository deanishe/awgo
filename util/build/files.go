// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

package build

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar"

	"github.com/deanishe/awgo/util"
)

// Export builds an .alfredworkflow file in directory dest
// from the files in directory src. If src is an empty string,
// "build" is used; if dest is empty, "dist" is used.
//
// The filename of the workflow file is generated automatically from
// the workflow's info.plist and is returned if zipping succeeds.
func Export(src, dest string) (path string, err error) {
	if src == "" {
		src = "build"
	}
	if dest == "" {
		dest = "dist"
	}
	var info *Info
	if info, err = NewInfo(InfoPlist(filepath.Join(src, "info.plist"))); err != nil {
		return
	}
	name := util.Slugify(fmt.Sprintf("%s-%s.alfredworkflow", info.Name, info.Version))
	if err = os.MkdirAll(dest, 0700); err != nil {
		return
	}
	path = filepath.Join(dest, name)

	if util.PathExists(path) {
		if err = os.Remove(path); err != nil {
			return
		}
	}

	var z *os.File
	if z, err = os.Create(path); err != nil {
		return
	}
	defer z.Close()

	out := zip.NewWriter(z)

	err = filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.IsDir() {
			return nil
		}

		var (
			name, orig string
			info       os.FileInfo
			mode       os.FileMode
		)

		if name, err = filepath.Rel(src, path); err != nil {
			return err
		}
		if orig, err = filepath.EvalSymlinks(path); err != nil {
			return err
		}
		if info, err = os.Stat(orig); err != nil {
			return err
		}
		mode = info.Mode()

		var (
			f  *os.File
			w  io.Writer
			fh = &zip.FileHeader{
				Name:   name,
				Method: zip.Deflate,
			}
		)
		fh.SetMode(mode.Perm())
		fh.SetModTime(info.ModTime())

		if f, err = os.Open(orig); err != nil {
			return err
		}
		defer f.Close()

		if w, err = out.CreateHeader(fh); err != nil {
			return err
		}
		if _, err = io.Copy(w, f); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return
	}

	err = out.Close()

	return
}

// Glob is a pattern and (relative) destination directory.
type Glob struct {
	// Pattern is a glob-style pattern to match against filesystem
	Pattern string
	// DestDir is a relative directory within target directory
	// to where files matching Pattern should be linked.
	DestDir string
}

// Globs creates a slice of Globs for patterns.
func Globs(pattern ...string) []Glob {
	globs := make([]Glob, len(pattern))

	for i, s := range pattern {
		globs[i] = Glob{Pattern: s}
	}

	return globs
}

// SymlinkGlobs symlinks multiple Globs to a directory.
func SymlinkGlobs(destDir string, globs ...Glob) error {
	for _, g := range globs {
		files, err := doublestar.Glob(g.Pattern)
		if err != nil {
			return err
		}

		for _, p := range files {
			dest := filepath.Join(destDir, g.DestDir, p)
			if err := Symlink(dest, p, true); err != nil {
				return err
			}
		}
	}
	return nil
}

// Symlink creates a symlink to target.
func Symlink(link, target string, relative bool) error {
	var (
		dir  string
		path string
		err  error
	)

	if link == "" {
		return errors.New("empty link")
	}

	if link, err = filepath.Abs(link); err != nil {
		return err
	}
	dir = filepath.Dir(link)

	if target, err = filepath.Abs(target); err != nil {
		return err
	}

	if _, err := os.Stat(target); err != nil {
		return err
	}

	if err = os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	path = target
	if relative {
		if path, err = filepath.Rel(dir, target); err != nil {
			return err
		}
	}

	fmt.Printf("%s  -->  %s\n", link, path)
	return os.Symlink(path, link)
}
