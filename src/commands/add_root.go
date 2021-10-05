package commands

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/proofrock/snapkup/util"
)

var arSql1 = "SELECT 1 FROM ROOTS WHERE PATH = ? LIMIT 1"
var arSql2 = "INSERT INTO ROOTS (PATH) VALUES (?)"

func AddRoot(bkpDir string, _toAdd *string) error {
	dbPath := bkpDir + "/" + util.DbFileName

	toAdd, errAbsolutizing := filepath.Abs(*_toAdd)
	if errAbsolutizing != nil {
		return errAbsolutizing
	}

	if _, errNotExists := os.Stat(dbPath); os.IsNotExist(errNotExists) {
		return fmt.Errorf("Database does not exists, initialize backup dir first (%s)", dbPath)
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

	rows, errQuerying := db.Query(arSql1, toAdd)
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

	if _, errExecing := tx.Exec(arSql2, toAdd); errExecing != nil {
		tx.Rollback() // error is not managed
		return errExecing
	}

	if errCommitting := tx.Commit(); errCommitting != nil {
		return errCommitting
	}

	println(fmt.Sprintf("Root correctly added (%s)", toAdd))

	return nil
}
