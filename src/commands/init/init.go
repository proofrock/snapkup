package init

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path"

	_ "github.com/mattn/go-sqlite3"

	"github.com/proofrock/snapkup/util"
)

var sqls = [5]string{
	`CREATE TABLE "PARAMS" (
		"KEY"	TEXT NOT NULL,
		"VALUE"	TEXT NOT NULL,
		PRIMARY KEY("KEY")
	)`,
	`CREATE TABLE "SNAPS" (
		"ID"		INTEGER NOT NULL,
		"TIMESTAMP"	INTEGER NOT NULL,
		"LABEL"		TEXT NOT NULL,
		PRIMARY KEY("ID")
	)`,
	`CREATE TABLE "ITEMS" (
		"PATH"		TEXT NOT NULL,
		"SNAP"		INTEGER NOT NULL,
		"HASH"		TEXT NOT NULL,
		"IS_DIR"	INTEGER NOT NULL,
		"MODE"		INTEGER NOT NULL,
		"MOD_TIME"	INTEGER NOT NULL,
		PRIMARY KEY("PATH", "SNAP")
	)`,
	`CREATE TABLE "BLOBS" (
		"HASH"			TEXT NOT NULL,
		"SIZE"			INTEGER NOT NULL,
		"BLOB_SIZE"		INTEGER NOT NULL,
		"IS_COMPRESSED"	INTEGER NOT NULL,
		"IS_ENCRYPTED"	INTEGER NOT NULL,
		PRIMARY KEY("HASH")
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

	dbPath := path.Join(bkpDir, util.DbFileName)

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

	iv := make([]byte, 16, 16)
	if _, errRandomizing := rand.Read(iv); errRandomizing != nil {
		return errRandomizing
	}

	if _, errExecing := tx.Exec("INSERT INTO PARAMS (KEY, VALUE) VALUES ('IV', ?)", iv); errExecing != nil {
		tx.Rollback()
		return errExecing
	}

	if errCommitting := tx.Commit(); errCommitting != nil {
		return errCommitting
	}

	hex := []rune("0123456789abcdef")
	for i := 0; i < len(hex); i++ {
		os.Mkdir(path.Join(bkpDir, string(hex[i])), fs.FileMode(0700))
	}

	println("Backup directory correctly initialized in ", bkpDir)

	return nil
}
