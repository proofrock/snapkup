package util

import "database/sql"

type dbThingy interface {
	Query(qry string, args ...interface{}) (*sql.Rows, error)
}

func GetRootsList(db dbThingy) ([]string, error) {
	var roots []string
	rows, errQuerying := db.Query("SELECT PATH FROM ROOTS ORDER BY PATH ASC")
	if errQuerying != nil {
		return nil, errQuerying
	}
	defer rows.Close()
	for rows.Next() {
		var root string
		if errScanning := rows.Scan(&root); errScanning != nil {
			return nil, errScanning
		}
		roots = append(roots, root)
	}
	if errClosingQry := rows.Err(); errClosingQry != nil {
		return nil, errClosingQry
	}
	return roots, nil
}
