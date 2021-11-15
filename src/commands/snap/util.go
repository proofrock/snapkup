package snap

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/proofrock/snapkup/model"
	"github.com/proofrock/snapkup/util"
)

func findSnap(modl *model.Model, snap int) int {
	for i, snp := range modl.Snaps {
		if snp.Id == snap {
			return i
		}
	}

	return -1
}

type fileNfo struct {
	FullPath     string
	IsDir        bool
	IsEmpty      bool
	Hash         string
	Name         string
	Size         int64
	LastModified int64
	Mode         fs.FileMode
}

func walkFSTree(roots []string, key []byte, doHash bool) (files []fileNfo, numFiles int, numDirs int) {
	for _, root := range roots {
		if froot, errStatsing := os.Stat(root); errStatsing != nil {
			util.PrintlnfErr("Error in Stat() of root: %v", errStatsing)
		} else if froot.IsDir() {
			filepath.Walk(root, func(path string, f os.FileInfo, errWalking error) error {
				if errWalking != nil {
					util.PrintlnfErr("Error walking fs tree: %v", errWalking)
				} else {
					hash := ""
					isEmpty := true
					if f.IsDir() {
						numDirs++
					} else {
						numFiles++
						if f.Size() > 0 {
							isEmpty = false
							if doHash {
								if _hash, errHashing := util.FileHash(path, key); errHashing != nil {
									util.PrintlnfErr("Error hashing file: %v", errHashing)
								} else {
									hash = _hash
								}
							}
						}
					}

					files = append(files, fileNfo{
						IsDir:        f.IsDir(),
						IsEmpty:      isEmpty,
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
			if hash, errHashing := util.FileHash(root, key); errHashing != nil {
				util.PrintlnfErr("Error hashing file: %v", errHashing)
			} else {
				files = append(files, fileNfo{
					IsDir:        false,
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
