package fileutils

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	fp "path/filepath"
)

// IsDir checks if path is an existing dir.
func IsDir(path string) bool {
	f, err := os.Stat(path)
	if err != nil {
		return false
	}

	return f.IsDir()
}

// IsFile checks if path is an existing dir.
func IsFile(path string) bool {
	f, err := os.Stat(path)
	if err != nil {
		return false
	}

	return !f.IsDir()
}

// DirIsEmpty checks if a directory is empty.
func DirIsEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}

	return false, err
}

// SameFile check whether file in pathA and pathB is the same.
func SameFile(pathA, pathB string) bool {
	// Get file stats. If error happened then files must be different.
	statA, err := os.Stat(pathA)
	if err != nil {
		return false
	}

	statB, err := os.Stat(pathB)
	if err != nil {
		return false
	}

	// Make sure mod time, size and mode are the same.
	return statA.ModTime().Equal(statB.ModTime()) &&
		statA.Size() == statB.Size() &&
		statA.Mode() == statB.Mode()
}

// CopyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode will be copied from the source and
// the copied data is synced/flushed to stable storage.
// Copied from https://gist.github.com/r0l1/92462b38df26839a3ca324697c8cba04
// by Roland Singer
func CopyFile(src, dst string) (err error) {
	// Open source file
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	// Create or truncate output file
	err = os.MkdirAll(fp.Dir(dst), os.ModePerm)
	if err != nil {
		return
	}

	out, err := os.Create(dst)
	if err != nil {
		return
	}

	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	// Copy content from input to output
	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	// Make output has same permission as input
	si, err := in.Stat()
	if err != nil {
		return
	}

	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	// Make output has same time
	inTime := si.ModTime()
	err = os.Chtimes(dst, inTime, inTime)
	return
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
// Symlinks are ignored and skipped.
// Copied from https://gist.github.com/r0l1/92462b38df26839a3ca324697c8cba04
// by Roland Singer
func CopyDir(src string, dst string, excludedPaths map[string]struct{}) (err error) {
	// Clean source and destination path
	src = fp.Clean(src)
	dst = fp.Clean(dst)

	// Make sure source is not excluded
	if _, excluded := excludedPaths[src]; excluded {
		return nil
	}

	// Make sure source is exist and a directory
	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	// Make sure destination is not exist
	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}

	if err == nil {
		return fmt.Errorf("destination already exists")
	}

	// Create destination directory
	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	// Fetch entries of current directory
	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	// Recursively copy each entry
	for _, entry := range entries {
		srcPath := fp.Join(src, entry.Name())
		dstPath := fp.Join(dst, entry.Name())

		if _, excluded := excludedPaths[srcPath]; excluded {
			continue
		}

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath, excludedPaths)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}
