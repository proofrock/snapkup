package commands

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/proofrock/snapkup/util"
)

var lrSql = "SELECT PATH FROM ROOTS ORDER BY PATH ASC"

func ListRoots(bkpDir string) error {
	dbPath := bkpDir + "/" + util.DbFileName

	if _, errNotExists := os.Stat(dbPath); os.IsNotExist(errNotExists) {
		return fmt.Errorf("Database does not exists, initialize backup dir first (%s)", dbPath)
	}

	db, errOpeningDb := sql.Open("sqlite3", dbPath)
	if errOpeningDb != nil {
		return errOpeningDb
	}
	defer db.Close()

	var roots []string
	rows, errQuerying := db.Query(lrSql)
	if errQuerying != nil {
		return errQuerying
	}
	defer rows.Close()
	for rows.Next() {
		var root string
		if errScanning := rows.Scan(&root); errScanning != nil {
			return errScanning
		}
		roots = append(roots, root)
	}
	if errClosingQry := rows.Err(); errClosingQry != nil {
		return errClosingQry
	}

	for _, root := range roots {
		fmt.Printf("%s\n", root)
	}

	return nil
}
