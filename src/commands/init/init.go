package init

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	"github.com/proofrock/snapkup/util"
)

var sqls = [4]string{
	`CREATE TABLE "FS" (
		"PATH"	TEXT NOT NULL,
		"LEN"	INTEGER NOT NULL,
		"PERMS"	INTEGER NOT NULL,
		"FROM_GEN"	INTEGER NOT NULL,
		"TO_GEN"	INTEGER NOT NULL,
		"ID"	INTEGER NOT NULL,
		"STORED_IN"	INTEGER NOT NULL,
		"OFFSET" INTEGER NOT NULL,
		PRIMARY KEY("PATH","LEN","PERMS")
	)`,
	`CREATE TABLE "SNAPS" (
		"ID"	INTEGER NOT NULL,
		"DATE"	INTEGER NOT NULL,
		PRIMARY KEY("ID")
	)`,
	`CREATE TABLE "PARAMS" (
		"KEY"	TEXT NOT NULL,
		"VAL"	TEXT,
		PRIMARY KEY("KEY")
	)`,
	`CREATE TABLE "ROOTS" (
		"PATH"	TEXT NOT NULL,
		PRIMARY KEY("PATH")
	)`,
}

func Init(bkpDir string) error {
	if isEmpty, err := util.IsEmpty(bkpDir); err != nil {
		return err
	} else if !isEmpty {
		return fmt.Errorf("Backup dir is not empty (%s)", bkpDir)
	}

	dbPath := bkpDir + "/" + util.DbFileName

	db, errOpeningDb := sql.Open("sqlite3", dbPath)
	if errOpeningDb != nil {
		return errOpeningDb
	}
	defer db.Close()

	tx, errBeginning := db.Begin()
	if errBeginning != nil {
		return errBeginning
	}

	for _, sql := range sqls {
		if _, errExecing := tx.Exec(sql); errExecing != nil {
			tx.Rollback() // error is not managed
			return errExecing
		}
	}

	if errCommitting := tx.Commit(); errCommitting != nil {
		return errCommitting
	}

	println(fmt.Sprintf("Backup directory correctly initialized in %s", bkpDir))

	return nil
}
