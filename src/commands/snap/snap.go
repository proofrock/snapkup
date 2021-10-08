package snap

import (
	"database/sql"
	"fmt"
	"path"
	"sort"
	"time"

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

	st1, errSt1 := db.Prepare("SELECT UID FROM ITEMS WHERE PATH = ? AND HASH = ?")
	if errSt1 != nil {
		return errSt1
	}
	defer st1.Close()
	st2, errSt2 := db.Prepare("INSERT INTO ITEMS (PATH, HASH, IS_DIR, UID) VALUES (?, ?, ?, ?)")
	if errSt2 != nil {
		return errSt2
	}
	defer st2.Close()
	st3, errSt3 := db.Prepare("INSERT INTO LNK_ITEM_SNAP (UID, SNAP, MODE, MOD_TIME) VALUES (?, ?, ?, ?)")
	if errSt3 != nil {
		return errSt3
	}
	defer st3.Close()

	tx, errBeginning := db.Begin()
	if errBeginning != nil {
		return errBeginning
	}

	roots, errGettingRoots := util.GetRootsList(tx)
	if errGettingRoots != nil {
		tx.Rollback()
		return errGettingRoots
	}

	files, numFiles, numDirs := util.WalkFSTree(roots)
	fmt.Printf("Found %d files and %d directories.\n", numFiles, numDirs)

	sort.Slice(files, func(i int, j int) bool { return files[i].FullPath < files[j].FullPath })

	snap, errRecSnap := recNewSnap(tx)
	if errRecSnap != nil {
		tx.Rollback()
		return errRecSnap
	}

	uid, errGetMaxUID := getMaxUID(tx)
	if errGetMaxUID != nil {
		tx.Rollback()
		return errGetMaxUID
	}

	st1tx := tx.Stmt(st1)
	defer st1tx.Close()
	st2tx := tx.Stmt(st2)
	defer st2tx.Close()
	st3tx := tx.Stmt(st3)
	defer st3tx.Close()

	var newFiles []util.FileNfo
	for _, file := range files {
		var uidToSet int
		row := st1tx.QueryRow(file.FullPath, file.Hash)
		if err1 := row.Scan(&uidToSet); err1 == sql.ErrNoRows {
			// new item
			uidToSet = uid
			uid++
			newFiles = append(newFiles, file)
			_, err2 := st2tx.Exec(file.FullPath, file.Hash, file.IsDir, uidToSet)
			if err2 != nil {
				tx.Rollback()
				return err2
			}
		} else if err1 != nil {
			tx.Rollback()
			return err1
		}

		_, err3 := st3tx.Exec(uidToSet, snap, uint32(file.Mode), file.LastModified)
		if err3 != nil {
			tx.Rollback()
			return err3
		}
	}

	fmt.Printf("%d new items and %d already present.\n", len(newFiles), len(files)-len(newFiles))

	for _, file := range newFiles {
		if file.IsDir == 1 {
			continue
		}

		pathDest := path.Join(bkpDir, file.Hash)

		if errCopying := util.CopyNotOverwrite(file.FullPath, pathDest); errCopying != nil {
			tx.Rollback()
			return errCopying
		}
	}

	if errCommitting := tx.Commit(); errCommitting != nil {
		return errCommitting
	}

	fmt.Printf("Snap %d correctly created\n", snap)

	return nil
}

func recNewSnap(tx *sql.Tx) (nuSnap int, errNewSnap error) {
	row := tx.QueryRow("SELECT COALESCE(MAX(ID) + 1, 0) FROM SNAPS")
	errNewSnap = row.Scan(&nuSnap)
	if errNewSnap != nil {
		return
	}
	_, errNewSnap = tx.Exec("INSERT INTO SNAPS (ID, TIMESTAMP) VALUES (?, ?)", nuSnap, time.Now().UnixMilli())
	return
}

func getMaxUID(tx *sql.Tx) (uid int, errGetMaxUID error) {
	row := tx.QueryRow("SELECT COALESCE(MAX(UID) + 1, 0) FROM ITEMS")
	errGetMaxUID = row.Scan(&uid)
	return
}
