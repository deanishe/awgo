// Copyright (c) 2019 Dean Jackson <deanishe@deanishe.net>
// MIT Licence applies http://opensource.org/licenses/MIT

package build

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func withTempDir(fn func(dir string)) {
	tmp, err := ioutil.TempDir("", "awgo-")
	if err != nil {
		panic(err)
	}

	path, err := filepath.EvalSymlinks(tmp)
	if err != nil {
		panic(err)
	}

	fn(path)

	defer func() {
		if err := os.RemoveAll(tmp); err != nil {
			panic(fmt.Sprintf("remove temp dir: %v", err))
		}
	}()
}

func TestSymlink(t *testing.T) {
	withTempDir(func(dir string) {
		tests := []struct {
			link     string
			target   string
			relative bool
			err      bool
		}{
			{"", "", true, true},
			{dir + "/dest.1.txt", "src.txt", true, true},
			{dir + "/info.plist", "./testdata/info.plist", true, false},
		}

		for _, td := range tests {
			td := td
			t.Run(fmt.Sprintf("link=%q, target=%q", td.link, td.target), func(t *testing.T) {
				t.Parallel()
				err := Symlink(td.link, td.target, td.relative)
				if td.err {
					if err == nil {
						t.Error("Expected error")
					}
					return
				}
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}

				if _, err := os.Stat(td.link); err != nil {
					t.Fatal(err)
				}

				p, err := filepath.EvalSymlinks(td.link)
				if err != nil {
					t.Fatal(err)
				}
				target, err := filepath.Abs(td.target)
				if err != nil {
					t.Fatal(err)
				}
				if p != target {
					t.Errorf("Bad Symlink. Expected=%q, Got=%q", target, p)
				}
			})
		}
	})
}

func TestGlobs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		pattern string
		files   []string
	}{
		{"testdata/workflow/*.plist", []string{
			"testdata/workflow/info.plist",
		}},
		{"testdata/workflow/*", []string{
			"testdata/workflow/info.plist",
			"testdata/workflow/script.sh",
			"testdata/workflow/icon.png",
		}},
	}

	for _, td := range tests {
		td := td
		t.Run(td.pattern, func(t *testing.T) {
			withTempDir(func(dir string) {
				t.Parallel()

				g := Globs(td.pattern)[0]
				if g.Pattern != td.pattern {
					t.Errorf("Bad Pattern. Expected=%q, Got=%q", td.pattern, g.Pattern)
				}

				if err := SymlinkGlobs(dir, g); err != nil {
					t.Fatal(err)
				}

				for _, p1 := range td.files {
					p2 := filepath.Join(dir, p1)
					if err := compareFiles(p1, p2); err != nil {
						t.Errorf("Bad File (%s): %v", p1, err)
					}
				}
			})
		})
	}
}

func TestExport(t *testing.T) {
	t.Parallel()

	for _, src := range []string{"./testdata/workflow", "./testdata/workflow-symlinked"} {
		src := src
		t.Run(src, func(t *testing.T) {
			env := map[string]string{
				"alfred_version":     "4.0.3",
				"alfred_preferences": "./testbuild",
			}
			withEnv(env, func() {
				withTempDir(func(dir string) {
					var (
						xdir = filepath.Join(dir, "extracted")
						path string
						err  error
					)
					if err = os.Mkdir(xdir, 0700); err != nil {
						t.Fatal(err)
					}
					if path, err = Export(src, dir); err != nil {
						t.Fatal(err)
					}

					if _, err = os.Stat(path); err != nil {
						t.Fatal(err)
					}

					name := filepath.Base(path)
					x := "AwGo-1.2.0.alfredworkflow"
					if name != x {
						t.Errorf("Bad Name. Expected=%q, Got=%q", x, name)
					}

					cmd := exec.Command("unzip", path, "-d", xdir)
					if err = cmd.Run(); err != nil {
						t.Fatal(err)
					}

					if err = compareDirs(src, xdir); err != nil {
						t.Fatal(err)
					}
				})
			})
		})
	}

}

type fileInfo struct {
	Name    string
	ModTime time.Time
	Mode    os.FileMode
	Size    int64
	Hash    string
}

func TestCompareDirs(t *testing.T) {
	t.Parallel()

	if err := compareDirs("./testdata/workflow", "./testdata/workflow-symlinked"); err != nil {
		t.Error(err)
	}
}

func fileStats(path string) (fileInfo, error) {
	var (
		info fileInfo
		fi   os.FileInfo
		err  error
	)
	if fi, err = os.Stat(path); err != nil {
		return info, err
	}
	info.Name = fi.Name()

	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return info, err
	}

	if fi, err = os.Stat(path); err != nil {
		return info, err
	}
	info.ModTime = fi.ModTime().Truncate(time.Second)
	info.Mode = fi.Mode()
	info.Size = fi.Size()

	hash, err := hashFile(path)
	if err != nil {
		return info, err
	}
	info.Hash = hash

	return info, nil
}

func compareDirs(dir1, dir2 string) error {
	var (
		files1, files2 []fileInfo
		err            error
	)
	read := func(dir string) ([]fileInfo, error) {
		var infos []fileInfo
		err := filepath.Walk(dir, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if fi.IsDir() {
				return nil
			}

			info, err := fileStats(path)
			if err != nil {
				return err
			}

			infos = append(infos, info)

			return nil
		})
		if err != nil {
			return nil, err
		}
		return infos, nil
	}

	if files1, err = read(dir1); err != nil {
		return fmt.Errorf("read dir %q: %v", dir1, err)
	}
	if files2, err = read(dir2); err != nil {
		return fmt.Errorf("read dir %q: %v", dir2, err)
	}

	if len(files1) != len(files2) {
		return fmt.Errorf("unequal sizes (%d vs %d)", len(files1), len(files2))
	}

	for i, f1 := range files1 {
		f2 := files2[i]
		if f1 != f2 {
			return fmt.Errorf("item %d unequal: (%v vs %v)", i, f1, f2)
		}
	}

	return nil
}

func compareFiles(path1, path2 string) error {
	var (
		info1, info2 fileInfo
		err          error
	)

	if info1, err = fileStats(path1); err != nil {
		return err
	}
	if info2, err = fileStats(path2); err != nil {
		return err
	}

	if info1 != info2 {
		return fmt.Errorf("unequal files (%v vs %v)", info1, info2)
	}

	return nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
