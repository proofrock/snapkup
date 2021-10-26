package util

import (
	"io"
	"os"

	"github.com/proofrock/snapkup/util/streams"
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

func Restore(src string, dst string, isCompressed bool) error {
	key := make([]byte, 32) // TODO implment

	if _, errStatsing := os.Stat(dst); !os.IsNotExist(errStatsing) {
		// an identical file already exists
		return nil
	}

	source, errOpening := os.Open(src)
	if errOpening != nil {
		return errOpening
	}
	defer source.Close()

	destination, errCreating := os.Create(dst)
	if errCreating != nil {
		return errCreating
	}
	defer destination.Close()

	ins, err := streams.NewIS(key, source)
	if err != nil {
		return err
	}
	defer ins.Close()

	if _, err = io.Copy(destination, ins); err != nil {
		return err
	}

	if err = ins.Close(); err != nil {
		return err
	}

	return nil
}
