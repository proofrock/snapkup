package util

import (
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/DataDog/zstd"
	"github.com/dchest/siphash"
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

type FileNfo struct {
	FullPath     string
	IsDir        int
	Hash         string
	Name         string
	Size         int64
	LastModified int64
	Mode         fs.FileMode
}

func WalkFSTree(roots []string, iv []byte) (files []FileNfo, numFiles int, numDirs int) {
	for _, root := range roots {
		if froot, errStatsing := os.Stat(root); errStatsing != nil {
			fmt.Fprintf(os.Stderr, "Error in Stat() of root: %v\n", errStatsing)
		} else if froot.IsDir() {
			filepath.Walk(root, func(path string, f os.FileInfo, errWalking error) error {
				if errWalking != nil {
					fmt.Fprintf(os.Stderr, "Error walking fs tree: %v\n", errWalking)
				} else {
					var hash string
					var isDir int
					if f.IsDir() {
						numDirs++
						isDir = 1
					} else {
						numFiles++
						isDir = 0
						if _hash, errHashing := FileHash(path, iv); errHashing != nil {
							fmt.Fprintf(os.Stderr, "Error hashing file: %v\n", errHashing)
						} else {
							hash = _hash
						}
					}

					files = append(files, FileNfo{
						IsDir:        isDir,
						FullPath:     path,
						Hash:         hash,
						Name:         f.Name(),
						Size:         f.Size(),
						LastModified: f.ModTime().Unix(),
						Mode:         f.Mode(),
					})
				}
				return nil
			})
		} else {
			if hash, errHashing := FileHash(root, iv); errHashing != nil {
				fmt.Fprintf(os.Stderr, "Error hashing file: %v\n", errHashing)
			} else {
				files = append(files, FileNfo{
					IsDir:        0,
					FullPath:     root,
					Hash:         hash,
					Name:         froot.Name(),
					Size:         froot.Size(),
					LastModified: froot.ModTime().Unix(),
					Mode:         froot.Mode(),
				})
				numFiles = 1
			}
		}
	}

	return
}

const bufSize = 1024 * 32 // 32Kb

func FileHash(path string, iv []byte) (string, error) {
	source, errOpening := os.Open(path)
	if errOpening != nil {
		return "", errOpening
	}
	defer source.Close()

	hasher := siphash.New128(iv)
	buf := make([]byte, bufSize)
	for {
		n, errHashingFile := source.Read(buf)
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

func simpleCopy(src string, dst string) error {
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

	_, errCopying := io.Copy(destination, source)
	if errCopying != nil {
		return errCopying
	}

	return nil
}

func Store(src string, dst string, compress bool) (blobSize int64, err error) {
	if compress {
		source, errOpening := os.Open(src)
		if errOpening != nil {
			err = errOpening
			return
		}
		defer source.Close()

		destination, errCreating := os.Create(dst)
		if errCreating != nil {
			err = errCreating
			return
		}
		defer destination.Close()

		zDestination := zstd.NewWriterLevel(destination, 9)
		defer zDestination.Close()

		if _, errCopyingStreams := io.Copy(zDestination, source); errCopyingStreams != nil {
			err = errCopyingStreams
			return
		}

		zDestination.Flush()
	} else {
		if errCopying := simpleCopy(src, dst); errCopying != nil {
			err = errCopying
			return
		}
	}

	stat, _ := os.Stat(dst)
	blobSize = stat.Size()

	return
}

func Restore(src string, dst string, isCompressed bool) error {
	if _, errStatsing := os.Stat(dst); !os.IsNotExist(errStatsing) {
		// an identical file already exists
		return nil
	}

	if !isCompressed {
		// it's not compressed. Simply copy and return.
		if errCopying := simpleCopy(src, dst); errCopying != nil {
			return errCopying
		}

		return nil
	}

	source, errOpening := os.Open(src)
	if errOpening != nil {
		return errOpening
	}
	defer source.Close()

	zSource := zstd.NewReader(source)
	defer zSource.Close()

	destination, errCreating := os.Create(dst)
	if errCreating != nil {
		return errCreating
	}
	defer destination.Close()

	_, errCopying := io.Copy(destination, zSource)
	if errCopying != nil {
		return errCopying
	}

	return nil
}
