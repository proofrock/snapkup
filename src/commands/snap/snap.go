package snap

import (
	"database/sql"

	"github.com/proofrock/snapkup/util"
)

func Snap(bkpDir string) error {
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

	files := util.WalkFSTree(roots)

	return nil
}
