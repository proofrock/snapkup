package listroots

import (
	"database/sql"

	"github.com/proofrock/snapkup/util"
)

func ListRoots(bkpDir string) error {
	dbPath, errComposingDbPath := util.DbFile(bkpDir)
	if errComposingDbPath != nil {
		return errComposingDbPath
	}

	db, errOpeningDb := sql.Open("sqlite3", dbPath)
	if errOpeningDb != nil {
		return errOpeningDb
	}
	defer db.Close()

	roots, errGettingRoots := util.GetRootsList(db)
	if errGettingRoots != nil {
		return errGettingRoots
	}

	for _, root := range roots {
		println(root)
	}

	return nil
}
