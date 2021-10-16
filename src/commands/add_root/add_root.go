package addroot

import (
	"database/sql"
	"fmt"

	"github.com/proofrock/snapkup/util"
)

func AddRoot(bkpDir string, toAdd string) error {
	dbPath, errComposingDbPath := util.DbFile(bkpDir)
	if errComposingDbPath != nil {
		return errComposingDbPath
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

	throwaway := 1
	row := db.QueryRow("SELECT 1 FROM ROOTS WHERE PATH = ?", toAdd)
	if errQuerying := row.Scan(&throwaway); errQuerying == nil {
		tx.Rollback()
		return fmt.Errorf("Root already present (%s)", toAdd)
	} else if errQuerying != sql.ErrNoRows {
		tx.Rollback()
		return errQuerying
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
