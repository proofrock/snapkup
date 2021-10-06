package delroot

import (
	"database/sql"
	"fmt"

	"github.com/proofrock/snapkup/util"
)

var drSql = "DELETE FROM ROOTS WHERE PATH = ?"

func DelRoot(bkpDir string, toDel *string) error {
	dbPath, errComposingDbPath := util.DbFile(bkpDir)
	if errComposingDbPath != nil {
		return errComposingDbPath
	}

	db, errOpeningDb := sql.Open("sqlite3", dbPath)
	if errOpeningDb != nil {
		return errOpeningDb
	}
	defer db.Close()

	if res, errExecing := db.Exec(drSql, *toDel); errExecing != nil {
		return errExecing
	} else if numAffected, errCalcRowsAffected := res.RowsAffected(); errCalcRowsAffected != nil {
		return errExecing
	} else if numAffected == 0 {
		return fmt.Errorf("Root not found in pool (%s)", *toDel)
	}

	println("Root correctly deleted (" + *toDel + ")")

	return nil
}
