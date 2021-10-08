package init

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	"github.com/proofrock/snapkup/util"
)

var sqls = [4]string{
	`CREATE TABLE "SNAPS" (
		"ID"		INTEGER NOT NULL,
		"TIMESTAMP"	INTEGER NOT NULL,
		PRIMARY KEY("ID")
	)`,
	`CREATE TABLE "ITEMS" (
		"PATH"		TEXT NOT NULL,
		"HASH"		TEXT NOT NULL,
		"IS_DIR"	INTEGER NOT NULL,
		"UID"		INTEGER NOT NULL UNIQUE,
		PRIMARY KEY("PATH", "HASH")
	)`,
	`CREATE TABLE "LNK_ITEM_SNAP" (
		"UID"		INTEGER NOT NULL,
		"SNAP"		INTEGER NOT NULL,
		"MODE"		TEXT NOT NULL,
		"MOD_TIME"	INTEGER NOT NULL,
		PRIMARY KEY("UID", "SNAP"),
		FOREIGN KEY("UID") REFERENCES ITEMS("UID"),
		FOREIGN KEY("SNAP") REFERENCES SNAPS("ID")
	)`,
	`CREATE TABLE "ROOTS" (
		"PATH"	TEXT NOT NULL,
		PRIMARY KEY("PATH")
	)`,
}

func Init(bkpDir string) error {
	if isEmpty, errCheckingEmpty := util.IsEmpty(bkpDir); errCheckingEmpty != nil {
		return errCheckingEmpty
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

	println("Backup directory correctly initialized in ", bkpDir)

	return nil
}
