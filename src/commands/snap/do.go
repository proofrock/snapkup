package snap

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/proofrock/snapkup/model"
	"github.com/proofrock/snapkup/util"
	"github.com/proofrock/snapkup/util/streams"
)

func Do(bkpDir string, dontCompress bool, label string) func(modl *model.Model) error {
	return func(modl *model.Model) error {
		files, numFiles, numDirs := walkFSTree(modl.Roots, modl.Key4Hashes)
		fmt.Printf("Found %d files and %d directories.\n", numFiles, numDirs)

		sort.Slice(files, func(i int, j int) bool { return files[i].FullPath < files[j].FullPath })

		// Find next snap (max + 1)
		var snap = 0
		for _, snp := range modl.Snaps {
			if snp.Id >= snap {
				snap = snp.Id + 1
			}
		}

		// Create and insert a new one
		modl.Snaps = append(modl.Snaps, model.Snap{Id: snap, Timestamp: time.Now().UnixMilli(), Label: label})

		// Extracts the existing hashes (blob ids)
		curHashes := make(map[string]bool)
		for _, blob := range modl.Blobs {
			curHashes[blob.Hash] = true
		}

		// Iterates over the items (files+dirs) found in the filesystem. Write them for
		// the new snap ID, and check if the corresponding blob is a duplicate of something
		// seen in the current scan, or from a previous scan. If so, the writing that
		// is performed is enough to create a reference.
		// In the end, the map newHashes contains the hashes that needs to be stored as a blob

		type finf struct {
			FullPath string
			Size     int64
		}

		newHashes := make(map[string]finf) // [hash]file_info
		for _, file := range files {
			modl.Items = append(modl.Items, model.Item{Path: file.FullPath, Snap: snap, Hash: file.Hash, IsDir: file.IsDir, Mode: int32(file.Mode.Perm()), ModTime: file.LastModified})

			if !file.IsDir {
				if !curHashes[file.Hash] {
					// hash not yet recorded, mark it for addition
					newHashes[file.Hash] = finf{file.FullPath, file.Size}
					curHashes[file.Hash] = true
				}
			}
		}

		fmt.Printf("%d new blobs to write\n", len(newHashes))

		if len(newHashes) > 0 {
			// Iterates over the blobs to write, and writes them (compressing or not)
			i := 1
			tot := len(newHashes)
			bar := pb.Full.Start(tot)
			for hash, finfo := range newHashes {
				pathDest := path.Join(bkpDir, hash[0:1], hash)

				bar.Increment()
				i++
				blobSize, errCopying := store(modl.Key4Enc, finfo.FullPath, pathDest, dontCompress)
				if errCopying != nil {
					return errCopying
				}

				modl.Blobs = append(modl.Blobs, model.Blob{Hash: hash, Size: finfo.Size, BlobSize: blobSize})
			}
			bar.Finish()
		}

		// TODO write this only after actual commit
		fmt.Printf("Snap %d correctly created\n", snap)

		return nil
	}
}

type fileNfo struct {
	FullPath     string
	IsDir        bool
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
					if f.IsDir() {
						numDirs++
					} else {
						numFiles++
						if _hash, errHashing := util.FileHash(path, key); errHashing != nil {
							fmt.Fprintf(os.Stderr, "Error hashing file: %v\n", errHashing)
						} else {
							hash = _hash
						}
					}

					files = append(files, fileNfo{
						IsDir:        f.IsDir(),
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
			if hash, errHashing := util.FileHash(root.Path, key); errHashing != nil {
				fmt.Fprintf(os.Stderr, "Error hashing file: %v\n", errHashing)
			} else {
				files = append(files, fileNfo{
					IsDir:        false,
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

func store(key []byte, src string, dst string, dontCompress bool) (blobSize int64, err error) {
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
