package list_snap

import (
	"database/sql"

	"github.com/proofrock/snapkup/util"
)

const sql1 = "SELECT PATH FROM ITEMS WHERE SNAP = ? ORDER BY PATH ASC"

func ListSnap(bkpDir string, snap int) error {
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
		var path string
		if errScanning := rows.Scan(&path); errScanning != nil {
			return errScanning
		}
		println(path)
	}
	if errClosingQry := rows.Err(); errClosingQry != nil {
		return errClosingQry
	}

	return nil
}
