package snap

import (
	"errors"
	"fmt"
	"github.com/proofrock/snapkup/util/agglos"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/proofrock/snapkup/model"
	"github.com/proofrock/snapkup/util"
	"github.com/proofrock/snapkup/util/streams"
)

func Restore(bkpDir string, snap int, restoreDir string, restorePrefixPath *string) func(modl *model.Model) error {
	return func(modl *model.Model) error {
		if findSnap(modl, snap) == -1 {
			return fmt.Errorf("snap %d not found in pool", snap)
		}

		if isEmpty, errCheckingEmpty := util.IsEmpty(restoreDir); errCheckingEmpty != nil {
			return errCheckingEmpty
		} else if !isEmpty {
			return fmt.Errorf("restore dir is not empty (%s)", restoreDir)
		}

		// finds files to restore
		numFiles := 0
		numDirs := 0
		var items []model.Item
		for _, itm := range modl.Items {
			if itm.Snap == snap && (restorePrefixPath != nil && strings.HasPrefix(itm.Path, *restorePrefixPath)) {
				items = append(items, itm)
				if itm.IsDir {
					numDirs++
				} else {
					numFiles++
				}
			}
		}

		fmt.Printf("Loaded %d files and %d directories.\n", numFiles, numDirs)

		blobs := make(map[string]model.Blob)
		for _, blob := range modl.Blobs {
			blobs[blob.Hash] = blob
		}

		for _, item := range items {
			dest := path.Join(restoreDir, removeDrive(item.Path))
			if !item.IsDir {
				// it's a file
				if errMkingDir := os.MkdirAll(filepath.Dir(dest), os.FileMode(0700)); errMkingDir != nil {
					return errMkingDir
				}

				if item.IsEmpty {
					if _, errCreatingEmptyFile := os.Create(dest); errCreatingEmptyFile != nil {
						return errCreatingEmptyFile
					}
				} else {
					blob := blobs[item.Hash]
					if blob.AggloRef == nil {
						source := path.Join(bkpDir, item.Hash[0:1], item.Hash)
						if errCopying := restore(modl.Key4Enc, source, dest); errCopying != nil {
							return errCopying
						}
					} else {
						source := path.Join(bkpDir, (*blob.AggloRef).AggloID[1:2], (*blob.AggloRef).AggloID)
						if errCopying := restoreFromAgglo(modl.Key4Enc, (*blob.AggloRef).Offset, blob.BlobSize, source, dest); errCopying != nil {
							return errCopying
						}
					}

					if !checkRestoredFile(dest, item.Hash, modl.Key4Hashes) {
						return errors.New(fmt.Sprintf("general checksum error in %s", dest))
					}
				}
			} else {
				if errMkingDir := os.MkdirAll(dest, os.FileMode(0700)); errMkingDir != nil {
					return errMkingDir
				}
			}
		}

		for _, item := range items {
			dest := path.Join(restoreDir, removeDrive(item.Path))

			if errChmod := os.Chmod(dest, fs.FileMode(item.Mode)); errChmod != nil {
				return errChmod
			}

			modTime := time.Unix(item.ModTime, 0)
			if errChtimes := os.Chtimes(dest, modTime, modTime); errChtimes != nil {
				return errChtimes
			}
		}

		fmt.Printf("Snapshot %d restored correctly.\n", snap)

		return nil
	}
}

func removeDrive(p string) string {
	if p[1] == ':' {
		return path.Join(p[0:1], p[2:])
	}
	return p
}

func restore(key []byte, src string, dst string) error {
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

func restoreFromAgglo(key []byte, offset, size int64, src string, dst string) error {
	if _, errStatsing := os.Stat(dst); !os.IsNotExist(errStatsing) {
		// an identical file already exists
		return nil
	}

	source, errOpening := os.Open(src)
	if errOpening != nil {
		return errOpening
	}
	defer source.Close()

	sourcePiece, errOpeningPiece := agglos.NewAIS(offset, size, source)
	if errOpeningPiece != nil {
		return errOpeningPiece
	}

	destination, errCreating := os.Create(dst)
	if errCreating != nil {
		return errCreating
	}
	defer destination.Close()

	ins, err := streams.NewIS(key, sourcePiece)
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

func checkRestoredFile(dest, recordedHash string, key []byte) bool {
	hash, errHashing := util.FileHash(dest, key)
	if errHashing != nil {
		return false
	}
	return hash == recordedHash
}
