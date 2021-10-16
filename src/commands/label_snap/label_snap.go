package labelsnap

import (
	"database/sql"

	"github.com/proofrock/snapkup/util"
)

func LabelSnap(bkpDir string, snap int, label string) error {
	dbPath, errComposingDbPath := util.DbFile(bkpDir)
	if errComposingDbPath != nil {
		return errComposingDbPath
	}

	db, errOpeningDb := sql.Open("sqlite3", dbPath)
	if errOpeningDb != nil {
		return errOpeningDb
	}
	defer db.Close()

	if _, errExecing := db.Exec("UPDATE SNAPS SET LABEL = ? WHERE ID = ?", label, snap); errExecing != nil {
		return errExecing
	}

	println("Ok.")

	return nil
}
