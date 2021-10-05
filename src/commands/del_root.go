package commands

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/proofrock/snapkup/util"
)

var drSql = "DELETE FROM ROOTS WHERE PATH = ?"

func DelRoot(bkpDir string, toDel *string) error {
	dbPath := bkpDir + "/" + util.DbFileName

	if _, errNotExists := os.Stat(dbPath); os.IsNotExist(errNotExists) {
		return fmt.Errorf("Database does not exists, initialize backup dir first (%s)", dbPath)
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

	println(fmt.Sprintf("Root correctly deleted (%s)", *toDel))

	return nil
}
