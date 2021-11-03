package snap

import (
	"fmt"
	"os"
	"path"

	"github.com/proofrock/snapkup/model"
)

func Delete(bkpDir string, toDel int) func(modl *model.Model) error {
	return func(modl *model.Model) error {
		var found = findSnap(modl, toDel)

		if found == -1 {
			return fmt.Errorf("snap %d not found in pool", toDel)
		}

		var nuItems []model.Item
		for _, item := range modl.Items {
			if item.Snap != toDel {
				nuItems = append(nuItems, item)
			}
		}
		modl.Items = nuItems

		modl.Snaps = append(modl.Snaps[:found], modl.Snaps[found+1:]...)

		fmt.Printf("Snap %d correctly deleted. Removing dangling files...\n", toDel)

		blobsToDel := make(map[string]bool)
		for _, blob := range modl.Blobs {
			blobsToDel[blob.Hash] = true
		}

		for _, item := range modl.Items {
			delete(blobsToDel, item.Hash)
		}

		fmt.Printf("Deleting %d blobs...\n", len(blobsToDel))
		for hash := range blobsToDel {
			pathToDel := path.Join(bkpDir, hash[0:1], hash)
			if errDeleting := os.Remove(pathToDel); errDeleting != nil {
				fmt.Fprintf(os.Stderr, "ERROR: deleting file %s; %v\n", hash, errDeleting)
			}
		}

		var nuBlobs []model.Blob
		for _, blob := range modl.Blobs {
			if _, isThere := blobsToDel[blob.Hash]; !isThere {
				nuBlobs = append(nuBlobs, blob)
			}
		}
		modl.Blobs = nuBlobs
		fmt.Printf("%d blobs remaining.\n", len(nuBlobs))

		println("All done.")

		return nil
	}
}

// TODO garbage collection
/*
func DelSnap(bkpDir string, toDel int) error {
	dbPath, errComposingDbPath := util.DbFile(bkpDir)
	if errComposingDbPath != nil {
		return errComposingDbPath
	}

	db, errOpeningDb := sql.Open("sqlite3", dbPath)
	if errOpeningDb != nil {
		return errOpeningDb
	}
	defer db.Close()

	tx, errBeginning := db.Begin()
	if errBeginning != nil {
		return errBeginning
	}

	if res, errExecing := tx.Exec("DELETE FROM SNAPS WHERE ID = ?", toDel); errExecing != nil {
		tx.Rollback()
		return errExecing
	} else if numAffected, errCalcRowsAffected := res.RowsAffected(); errCalcRowsAffected != nil {
		tx.Rollback()
		return errExecing
	} else if numAffected == 0 {
		tx.Rollback()
		return fmt.Errorf("Snap %d not found in pool", toDel)
	}

	if _, errExecing := tx.Exec("DELETE FROM ITEMS WHERE SNAP = ?", toDel); errExecing != nil {
		tx.Rollback()
		return errExecing
	}

	numDeleted := 0
	rows, errQuerying := tx.Query("SELECT HASH FROM BLOBS WHERE HASH NOT IN (SELECT HASH FROM ITEMS)")
	if errQuerying != nil {
		return errQuerying
	}
	defer rows.Close()
	for rows.Next() {
		var hash string
		if errScanning := rows.Scan(&hash); errScanning != nil {
			return errScanning
		}
		pathToDel := path.Join(bkpDir, hash[0:1], hash)
		if errDeleting := os.Remove(pathToDel); errDeleting != nil {
			fmt.Fprintf(os.Stderr, "ERROR: deleting file %s; %v\n", hash, errDeleting)
		}
		numDeleted++
	}
	if errClosingQry := rows.Err(); errClosingQry != nil {
		return errClosingQry
	}

	if _, errExecing := tx.Exec("DELETE FROM BLOBS WHERE HASH NOT IN (SELECT HASH FROM ITEMS)"); errExecing != nil {
		tx.Rollback()
		return errExecing
	}

	if errCommitting := tx.Commit(); errCommitting != nil {
		return errCommitting
	}

	fmt.Printf("Snap correctly deleted (%d); %d unnecessary files deleted.\n", toDel, numDeleted)

	return nil
}
*/
