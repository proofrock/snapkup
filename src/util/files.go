package util

import (
	"io"
	"os"
)

// Checked; from https://stackoverflow.com/questions/30697324/how-to-check-if-directory-on-path-is-empty
func IsEmpty(name string) (bool, error) {
	f, errOpening := os.Open(name)
	if errOpening != nil {
		return false, errOpening
	}
	defer f.Close()

	_, errReadingDir := f.Readdir(1)
	if errReadingDir == io.EOF {
		return true, nil
	}
	return false, errReadingDir
}
