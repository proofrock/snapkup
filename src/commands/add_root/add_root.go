package addroot

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/proofrock/snapkup/util"
)

func AddRoot(bkpDir string, _toAdd *string) error {
	dbPath, errComposingDbPath := util.DbFile(bkpDir)
	if errComposingDbPath != nil {
		return errComposingDbPath
	}

	toAdd, errAbsolutizing := filepath.Abs(*_toAdd)
	if errAbsolutizing != nil {
		return errAbsolutizing
	}

	db, errOpeningDb := sql.Open("sqlite3", dbPath)
	if errOpeningDb != nil {
		return errOpeningDb
	}
	defer db.Close()

	tx, errBeginning := db.Begin()
	if errBeginning != nil {
		return errBeginning
	}

	// TODO QueryOnce
	rows, errQuerying := db.Query("SELECT 1 FROM ROOTS WHERE PATH = ? LIMIT 1", toAdd)
	if errQuerying != nil {
		return errQuerying
	}
	defer rows.Close()
	if rows.Next() {
		tx.Rollback() // error is not managed
		return fmt.Errorf("Root already present (%s)", toAdd)
	}
	if errClosingQry := rows.Err(); errClosingQry != nil {
		return errClosingQry
	}

	if _, errExecing := tx.Exec("INSERT INTO ROOTS (PATH) VALUES (?)", toAdd); errExecing != nil {
		tx.Rollback() // error is not managed
		return errExecing
	}

	if errCommitting := tx.Commit(); errCommitting != nil {
		return errCommitting
	}

	println("Root correctly added (", toAdd, ")")

	return nil
}
