package cmd

import (
	"io"
	"os"
)

// dirEmpty checks if a directory is empty or not.
func dirEmpty(dirPath string) bool {
	dir, err := os.Open(dirPath)
	if err != nil {
		return false
	}
	defer dir.Close()

	_, err = dir.Readdirnames(1)
	if err != io.EOF {
		return false
	}

	return true
}

func panicError(err error, prefixes ...string) {
	if err != nil {
		cError.Print(prefixes)
		cError.Println(err)
		os.Exit(1)
	}
}
