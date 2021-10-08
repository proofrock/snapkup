package util

import (
	"fmt"
	"os"
	"path"
)

const DbFileName = "snapkup.db"

func DbFile(bkpDir string) (string, error) {
	dbPath := path.Join(bkpDir, DbFileName)

	if _, errNotExists := os.Stat(dbPath); os.IsNotExist(errNotExists) {
		return "", fmt.Errorf("Database does not exists, initialize backup dir first (%s)", dbPath)
	}

	return dbPath, nil
}
