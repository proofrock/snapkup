package util

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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
	isDir        bool
	FullPath     string
	Name         string
	Size         int64
	LastModified int64
	Mode         fs.FileMode
}

func WalkFSTree(roots []string) []FileNfo {
	var ret []FileNfo
	for _, root := range roots {
		filepath.Walk(root, func(path string, f os.FileInfo, errWalking error) error {
			if errWalking != nil {
				fmt.Fprintf(os.Stderr, "Error walking fs tree: %v\n", errWalking)
			} else {
				ret = append(ret, FileNfo{
					isDir:        f.IsDir(),
					FullPath:     path,
					Name:         f.Name(),
					Size:         f.Size(),
					LastModified: f.ModTime().Unix(),
					Mode:         f.Mode(),
				})
			}
			return nil
		})
	}
	return ret
}
