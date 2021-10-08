package restore

import (
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path"
	"time"

	"github.com/proofrock/snapkup/util"
)

const sql1 = `
SELECT PATH, HASH, MODE, MOD_TIME
  FROM ITEMS i
  JOIN LNK_ITEM_SNAP lis ON lis.UID = i.UID
 WHERE lis.SNAP = ?
 ORDER BY PATH ASC`

type item struct {
	Path    string
	Hash    string
	Mode    uint32
	ModTime int64
}

func Restore(bkpDir string, snap int, restoreDir string) error {
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
			if errScanning := rows.Scan(&item.Path, &item.Hash, &item.Mode, &item.ModTime); errScanning != nil {
				return errScanning
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
		if item.Hash == "" {
			// it's a dir
			os.MkdirAll(dest, os.ModePerm)
		} else {
			// it's a file
			source := path.Join(bkpDir, item.Hash)

			if errCopying := util.CopyNotOverwrite(source, dest); errCopying != nil {
				return errCopying
			}

			if errChmod := os.Chmod(dest, fs.FileMode(item.Mode)); errChmod != nil {
				return errChmod
			}
			if errChtimes := os.Chtimes(dest, time.Unix(item.ModTime, 0), time.Unix(item.ModTime, 0)); errChtimes != nil {
				return errChtimes
			}
		}
	}

	fmt.Printf("Snapshot %d restored correctly.\n", snap)

	return nil
}
