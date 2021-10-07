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
)

// Checked; from https://stackoverflow.com/questions/30697324/how-to-check-if-directory-on-path-is-empty
func IsEmpty(name string) (bool, error) {
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

type FileNfo struct {
	IsDir        int
	FullPath     string
	Hash         string
	Name         string
	Size         int64
	LastModified int64
	Mode         fs.FileMode
}

func WalkFSTree(roots []string) (files []FileNfo, numFiles int, numDirs int) {
	for _, root := range roots {
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
					if _hash, errHashing := FileHash(path); errHashing != nil {
						fmt.Fprintf(os.Stderr, "Error walking fs tree: %v\n", errWalking)
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
	}

	return
}

const bufSize = 1024 * 32 // 32Kb
var nothingUpMySleeve = []byte("SnapkupIsCool!!!")

func FileHash(path string) (string, error) {
	source, errOpening := os.Open(path)
	if errOpening != nil {
		return "", errOpening
	}
	defer source.Close()

	hasher := siphash.New128(nothingUpMySleeve)
	buf := make([]byte, bufSize)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return "", err
		}
		if n == 0 {
			break
		}

		if _, err := hasher.Write(buf[:n]); err != nil {
			return "", err
		}
	}

	ret := hasher.Sum(nil)
	stat, _ := source.Stat()
	len := stat.Size()

	for len > 0 {
		ret = append(ret, byte(0xff&len))
		len >>= 8
	}

	return strings.ToLower(hex.EncodeToString(ret)), nil
}