package util

import (
	"encoding/hex"
	"github.com/dchest/siphash"
	"io"
	"os"
	"strings"
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

const bufSize = 1024 * 32 // 32Kb

func FileHash(path string, key []byte) (string, error) {
	source, errOpening := os.Open(path)
	if errOpening != nil {
		return "", errOpening
	}
	defer source.Close()

	return DataHash(source, key)
}

func DataHash(reader io.Reader, key []byte) (string, error) {
	hasher := siphash.New128(key)
	buf := make([]byte, bufSize)
	for {
		n, errHashingFile := reader.Read(buf)
		if errHashingFile != nil && errHashingFile != io.EOF {
			return "", errHashingFile
		}
		if n == 0 {
			break
		}

		if _, errWritingHash := hasher.Write(buf[:n]); errWritingHash != nil {
			return "", errWritingHash
		}
	}

	ret := hasher.Sum(nil)

	return strings.ToLower(hex.EncodeToString(ret)), nil
}
