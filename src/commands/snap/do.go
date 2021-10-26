package snap

import (
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dchest/siphash"
	"github.com/proofrock/snapkup/model"
	"github.com/proofrock/snapkup/util"
	"github.com/proofrock/snapkup/util/streams"
)

func Do(dontCompress bool, label string) func(modl *model.Model) error {
	return func(modl *model.Model) error {
		files, numFiles, numDirs := walkFSTree(modl.Roots, modl.Key4Hashes)
		fmt.Printf("Found %d files and %d directories.\n", numFiles, numDirs)

		sort.Slice(files, func(i int, j int) bool { return files[i].FullPath < files[j].FullPath })

		// Find next snap (max + 1)
		var snap uint32 = 0
		for _, snp := range modl.Snaps {
			if snp.Id >= snap {
				snap = snp.Id + 1
			}
		}

		// Create and insert a new one
		modl.Snaps = append(modl.Snaps, model.Snap{Id: snap, Timestamp: time.Now().UnixMilli(), Label: label})

		return nil
	}
}

type fileNfo struct {
	FullPath     string
	IsDir        int
	Hash         string
	Name         string
	Size         int64
	LastModified int64
	Mode         fs.FileMode
}

func walkFSTree(roots []model.Root, key []byte) (files []fileNfo, numFiles int, numDirs int) {
	for _, root := range roots {
		if froot, errStatsing := os.Stat(root.Path); errStatsing != nil {
			fmt.Fprintf(os.Stderr, "Error in Stat() of root: %v\n", errStatsing)
		} else if froot.IsDir() {
			filepath.Walk(root.Path, func(path string, f os.FileInfo, errWalking error) error {
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
						if _hash, errHashing := fileHash(path, key); errHashing != nil {
							fmt.Fprintf(os.Stderr, "Error hashing file: %v\n", errHashing)
						} else {
							hash = _hash
						}
					}

					files = append(files, fileNfo{
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
			if hash, errHashing := fileHash(root.Path, key); errHashing != nil {
				fmt.Fprintf(os.Stderr, "Error hashing file: %v\n", errHashing)
			} else {
				files = append(files, fileNfo{
					IsDir:        0,
					FullPath:     root.Path,
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

func fileHash(path string, key []byte) (string, error) {
	source, errOpening := os.Open(path)
	if errOpening != nil {
		return "", errOpening
	}
	defer source.Close()

	hasher := siphash.New128(key)
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

func store(src string, dst string, dontCompress bool) (blobSize int64, err error) {
	key := make([]byte, 32) // TODO implment

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

	ous, err := streams.NewOS(key, util.ChunkSize, dontCompress, destination)
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
