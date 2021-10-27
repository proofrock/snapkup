package snap

import (
	"io"
	"os"

	"github.com/proofrock/snapkup/model"
	"github.com/proofrock/snapkup/util/streams"
)

func Restore(snap int, restoreDir string, restorePrefixPath *string) func(modl *model.Model) error {
	return func(modl *model.Model) error {
		return nil
	}
}

func restore(src string, dst string, isCompressed bool) error {
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

/*
const sql1 = `
SELECT i.PATH, i.HASH, i.IS_DIR, COALESCE(b.IS_COMPRESSED, 0), i.MODE, i.MOD_TIME
  FROM ITEMS i
  LEFT JOIN BLOBS b ON b.HASH = i.HASH
 WHERE i.SNAP = ?
 ORDER BY i.PATH ASC`

type item struct {
	Path         string
	Hash         string
	IsDir        int
	IsCompressed int
	Mode         uint32
	ModTime      int64
}

func Restore(bkpDir string, snap int, restoreDir string, restorePrefixPath *string) error {
	if isEmpty, errCheckingEmpty := util.IsEmpty(restoreDir); errCheckingEmpty != nil {
		return errCheckingEmpty
	} else if !isEmpty {
		return fmt.Errorf("Restore dir is not empty (%s)", restoreDir)
	}

	numFiles := 0
	numDirs := 0
	var items []item
	{
		dbPath, errComposingDbPath := util.DbFile(bkpDir)
		if errComposingDbPath != nil {
			return errComposingDbPath
		}

		db, errOpeningDb := sql.Open("sqlite3", dbPath)
		if errOpeningDb != nil {
			return errOpeningDb
		}
		defer db.Close()

		rows, errQuerying := db.Query(sql1, snap)
		if errQuerying != nil {
			return errQuerying
		}
		defer rows.Close()
		for rows.Next() {
			var item item
			if errScanning := rows.Scan(&item.Path, &item.Hash, &item.IsDir, &item.IsCompressed, &item.Mode, &item.ModTime); errScanning != nil {
				return errScanning
			}

			if restorePrefixPath != nil && !strings.HasPrefix(item.Path, *restorePrefixPath) {
				continue
			}

			if item.Hash == "" {
				numDirs++
			} else {
				numFiles++
			}

			items = append(items, item)
		}
		if errClosingQry := rows.Err(); errClosingQry != nil {
			return errClosingQry
		}

		fmt.Printf("Loaded %d files and %d directories.\n", numFiles, numDirs)
	}

	for _, item := range items {
		dest := path.Join(restoreDir, item.Path)
		if item.IsDir == 0 {
			// it's a file
			source := path.Join(bkpDir, item.Hash[0:1], item.Hash)

			if errMkingDir := os.MkdirAll(filepath.Dir(dest), os.FileMode(0700)); errMkingDir != nil {
				return errMkingDir
			}

			if errCopying := util.Restore(source, dest, item.IsCompressed == 1); errCopying != nil {
				return errCopying
			}
		} else {
			if errMkingDir := os.MkdirAll(dest, os.FileMode(0700)); errMkingDir != nil {
				return errMkingDir
			}
		}
	}

	for _, item := range items {
		dest := path.Join(restoreDir, item.Path)

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
*/
