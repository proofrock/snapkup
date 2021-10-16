package listsnaps

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/proofrock/snapkup/util"
)

type snap struct {
	id        int
	timestamp int64
	label     string
}

func ListSnaps(bkpDir string) error {
	dbPath, errComposingDbPath := util.DbFile(bkpDir)
	if errComposingDbPath != nil {
		return errComposingDbPath
	}

	db, errOpeningDb := sql.Open("sqlite3", dbPath)
	if errOpeningDb != nil {
		return errOpeningDb
	}
	defer db.Close()

	var snaps []snap
	rows, errQuerying := db.Query("SELECT ID, TIMESTAMP, LABEL FROM SNAPS ORDER BY ID DESC")
	if errQuerying != nil {
		return errQuerying
	}
	defer rows.Close()
	for rows.Next() {
		var snap snap
		if errScanning := rows.Scan(&snap.id, &snap.timestamp, &snap.label); errScanning != nil {
			return errScanning
		}
		snaps = append(snaps, snap)
	}
	if errClosingQry := rows.Err(); errClosingQry != nil {
		return errClosingQry
	}

	for _, snap := range snaps {
		ts := time.UnixMilli(snap.timestamp).Local().Format("2 Jan 2006, 15:04:05 (MST)")
		fmt.Printf("Snap %d:\t%s\t%s\n", snap.id, ts, snap.label)
	}

	return nil
}
