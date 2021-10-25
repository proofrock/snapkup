package util

import (
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/dchest/siphash"
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

const chunkSize = 32 * 1024 * 1024

func Store(src string, dst string, dontCompress bool) (blobSize int64, err error) {
	key := make([]byte, 32)   // TODO implment
	compress := !dontCompress // TODO invert upstream

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

	ous, err := streams.NewOS(key, chunkSize, compress, destination)
	if err != nil {
		return 0, err
	}
	defer ous.Close()

	if _, err = io.Copy(ous, source); err != nil {
		return 0, err
	}

	if err = ous.Close(); err != nil {
		return 0, err
	}

	if err = destination.Close(); err != nil {
		return 0, err
	}

	stat, _ := os.Stat(dst)
	blobSize = stat.Size()

	return
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
