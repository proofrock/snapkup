package snap

import (
	"database/sql"
	"fmt"
	"path"
	"sort"
	"time"

	"github.com/proofrock/snapkup/util"

	pb "github.com/cheggaaa/pb/v3"
)

func Snap(bkpDir string, dontCompress bool, label string) error {
	dbPath, errComposingDbPath := util.DbFile(bkpDir)
	if errComposingDbPath != nil {
		return errComposingDbPath
	}

	db, errOpeningDb := sql.Open("sqlite3", dbPath)
	if errOpeningDb != nil {
		return errOpeningDb
	}
	defer db.Close()

	st1, errSt1 := db.Prepare("INSERT INTO ITEMS (PATH, SNAP, HASH, IS_DIR, MODE, MOD_TIME) VALUES (?, ?, ?, ?, ?, ?)")
	if errSt1 != nil {
		return errSt1
	}
	defer st1.Close()
	st2, errSt2 := db.Prepare("SELECT 1 FROM BLOBS WHERE HASH = ?")
	if errSt2 != nil {
		return errSt2
	}
	defer st2.Close()
	st3, errSt3 := db.Prepare("INSERT INTO BLOBS (HASH, SIZE, BLOB_SIZE, IS_COMPRESSED, IS_ENCRYPTED) VALUES (?, ?, ?, ?, ?)")
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

	var iv []byte
	row := db.QueryRow("SELECT VALUE FROM PARAMS WHERE KEY = 'IV'")
	if errQuerying := row.Scan(&iv); errQuerying != nil {
		tx.Rollback()
		return errQuerying
	}

	files, numFiles, numDirs := util.WalkFSTree(roots, iv)
	fmt.Printf("Found %d files and %d directories.\n", numFiles, numDirs)

	sort.Slice(files, func(i int, j int) bool { return files[i].FullPath < files[j].FullPath })

	var snap int
	row = tx.QueryRow("SELECT COALESCE(MAX(ID) + 1, 0) FROM SNAPS")
	if errIdSnap := row.Scan(&snap); errIdSnap != nil {
		tx.Rollback()
		return errIdSnap
	}
	_, errNewSnap := tx.Exec("INSERT INTO SNAPS (ID, TIMESTAMP, LABEL) VALUES (?, ?, ?)", snap, time.Now().UnixMilli(), label)
	if errNewSnap != nil {
		tx.Rollback()
		return errNewSnap
	}

	st1tx := tx.Stmt(st1)
	defer st1tx.Close()
	st2tx := tx.Stmt(st2)
	defer st2tx.Close()
	st3tx := tx.Stmt(st3)
	defer st3tx.Close()

	// Iterates over the items (files+dirs) found in the filesystem. Write them for
	// the new snap ID, and check if the corresponding blob is a duplicate of something
	// seen in the current scan, or from a previous scan. If so, the writing that
	// is performed is enough to create a reference.
	// In the end, the map newHashes contains the hashes that needs to be stored as a blob

	type finf struct {
		FullPath string
		Size     int64
	}

	newHashes := make(map[string]finf) // [hash]file_info
	for _, file := range files {
		if _, errInsertingItem := st1tx.Exec(file.FullPath, snap, file.Hash, file.IsDir, file.Mode, file.LastModified); errInsertingItem != nil {
			tx.Rollback()
			return errInsertingItem
		}

		if file.IsDir == 0 {
			if _, alreadySeen := newHashes[file.Hash]; !alreadySeen {
				//check if the hash is already in the blobs
				throwaway := 1
				row := st2tx.QueryRow(file.Hash)
				if errQuerying := row.Scan(&throwaway); errQuerying == sql.ErrNoRows {
					// hash not yet recorded, mark it for addition
					newHashes[file.Hash] = finf{file.FullPath, file.Size}
				} else if errQuerying != nil {
					tx.Rollback()
					return errQuerying
				}
			}
		}
	}

	fmt.Printf("%d new blobs to write\n", len(newHashes))

	// Iterates over the blobs to write, and writes them (compressing or not)
	i := 1
	tot := len(newHashes)
	bar := pb.Full.Start(tot)
	for hash, finfo := range newHashes {
		pathDest := path.Join(bkpDir, hash[0:1], hash)

		bar.Increment()
		i++
		blobSize, errCopying := util.Store(finfo.FullPath, pathDest, dontCompress)
		if errCopying != nil {
			tx.Rollback()
			return errCopying
		}

		iCompressed := 1
		if dontCompress {
			iCompressed = 0
		}

		if _, errInsertingBlob := st3tx.Exec(hash, finfo.Size, blobSize, iCompressed, 0); errInsertingBlob != nil {
			tx.Rollback()
			return errInsertingBlob
		}
	}
	bar.Finish()

	if errCommitting := tx.Commit(); errCommitting != nil {
		return errCommitting
	}

	fmt.Printf("Snap %d correctly created\n", snap)

	return nil
}
