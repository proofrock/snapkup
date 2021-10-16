package info_snap

import (
	"database/sql"
	"fmt"

	"github.com/proofrock/snapkup/util"
)

const sql1 = `WITH 
consts AS (SELECT ? AS snap),
data AS (
          SELECT 1 AS key, COUNT(1) AS val FROM ITEMS WHERE SNAP = (SELECT snap FROM consts) AND IS_DIR = 0
UNION ALL SELECT 2 AS key, COUNT(1) AS val FROM ITEMS WHERE SNAP = (SELECT snap FROM consts) AND IS_DIR = 1
UNION ALL SELECT 3 AS key, SUM(b.SIZE) AS val FROM ITEMS i, BLOBS b WHERE i.HASH = b.HASH AND SNAP = (SELECT snap FROM consts) AND IS_DIR = 0
UNION ALL SELECT 4 AS key, SUM(BLOB_SIZE) AS val FROM BLOBS WHERE HASH IN (SELECT HASH FROM ITEMS WHERE SNAP = (SELECT snap FROM consts) AND IS_DIR = 0)
UNION ALL SELECT 5 AS key, SUM(BLOB_SIZE) AS val FROM BLOBS
UNION ALL SELECT 6 AS key, SUM(BLOB_SIZE) AS val FROM BLOBS WHERE HASH NOT IN (SELECT HASH FROM ITEMS WHERE SNAP != (SELECT snap FROM consts) AND IS_DIR = 0)
)
SELECT val FROM data ORDER BY key ASC`

var titles = [6]string{
	"Files",
	"Directories",
	"Size",
	"Stored size",
	"Tot. stored (all snaps)",
	"Free when deleted",
}

var isInByte = [6]bool{false, false, true, true, true, true}

const suffixes = "KMGTPE"

const unit = 1024

func fmtBytes(b int64) string {
	if b < unit {
		return fmt.Sprintf("%d b", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cb",
		float64(b)/float64(div), suffixes[exp])
}

func InfoSnap(bkpDir string, snap int) error {
	maxLen := 0
	for _, title := range titles {
		if len(title) > maxLen {
			maxLen = len(title)
		}
	}

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
	i := 0
	for rows.Next() {
		var val int64
		if errScanning := rows.Scan(&val); errScanning != nil {
			return errScanning
		}
		if isInByte[i] {
			fmt.Printf(fmt.Sprintf("%%-%ds: %%s\n", maxLen), titles[i], fmtBytes(val))
		} else {
			fmt.Printf(fmt.Sprintf("%%-%ds: %%d\n", maxLen), titles[i], val)
		}
		i++
	}
	if errClosingQry := rows.Err(); errClosingQry != nil {
		return errClosingQry
	}

	return nil
}
