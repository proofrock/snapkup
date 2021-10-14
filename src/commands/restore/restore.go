package restore

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/proofrock/snapkup/util"
)

const sql1 = `
SELECT i.PATH, i.HASH, COALESCE(b.IS_COMPRESSED, 0), i.MODE, i.MOD_TIME
  FROM ITEMS i
  LEFT JOIN BLOBS b ON b.HASH = i.HASH
 WHERE i.SNAP = ?
 ORDER BY i.PATH ASC`

type item struct {
	Path         string
	Hash         string
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
			if errScanning := rows.Scan(&item.Path, &item.Hash, &item.IsCompressed, &item.Mode, &item.ModTime); errScanning != nil {
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
		if item.Hash != "" {
			// it's a file
			source := path.Join(bkpDir, item.Hash[0:2], item.Hash[2:])

			if errMkingDir := os.MkdirAll(filepath.Dir(dest), os.FileMode(0700)); errMkingDir != nil {
				return errMkingDir
			}

			if errCopying := util.Restore(source, dest, item.IsCompressed == 1); errCopying != nil {
				return errCopying
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
