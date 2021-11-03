package snap

import (
	"fmt"
	"github.com/proofrock/snapkup/model"
	"github.com/proofrock/snapkup/util"
	"github.com/proofrock/snapkup/util/agglos"
	"github.com/proofrock/snapkup/util/streams"
	"os"
	"path"
)

func Check(bkpDir string) func(modl *model.Model) error {
	return func(modl *model.Model) error {
		allItems := make(map[string]bool)
		if len(modl.Agglos) > 0 {
			println("Checking agglos... ")
			for _, agglo := range modl.Agglos {
				ckAgglo(modl, bkpDir, agglo)
				allItems[agglo.ID] = true
			}
			println("Done.")
		}

		println("Checking blobs...")
		for _, blob := range modl.Blobs {
			if blob.AggloRef == nil { // normal blob
				ckNormalBlob(modl, bkpDir, blob)
			} else { // blob stored in an agglo
				if allItems[(*blob.AggloRef).AggloID] {
					ckBlobInAgglo(modl, bkpDir, blob)
				} else {
					fmt.Fprintf(os.Stderr, "ERROR: agglo %s not found for blob %s\n", (*blob.AggloRef).AggloID, blob.Hash)
				}
			}
			allItems[blob.Hash] = true
		}
		println("Done.")

		println("Checking filesystem...")
		files, _, _ := walkFSTree([]string{bkpDir}, nil, false)
		for _, file := range files {
			if file.IsDir || file.Name == model.ModelFileName {
				continue
			}
			if !allItems[file.Name] {
				fmt.Fprintf(os.Stderr, "WARNING: spurious file in filesystem: %s\n", file.FullPath)
			}
		}
		println("Done.")

		return nil
	}
}

func ckAgglo(modl *model.Model, bkpDir string, agglo model.Agglo) {
	fpath := path.Join(bkpDir, agglo.ID[1:2], agglo.ID)
	hash, errHashing := util.FileHash(fpath, modl.Key4Hashes)
	if errHashing != nil {
		fmt.Fprintf(os.Stderr, "ERROR: hashing agglo %s; %v\n", agglo.ID, errHashing)
	} else if hash != agglo.Hash {
		fmt.Fprintf(os.Stderr, "ERROR: bad checksum for agglo %s; %v\n", agglo.ID, errHashing)
	}
}

func ckNormalBlob(modl *model.Model, bkpDir string, blob model.Blob) {
	source, errOpening := os.Open(path.Join(bkpDir, blob.Hash[0:1], blob.Hash))
	if errOpening != nil {
		fmt.Fprintf(os.Stderr, "ERROR: opening blob %s; %v\n", blob.Hash, errOpening)
		return
	}
	defer source.Close()

	is, errInputStream := streams.NewIS(modl.Key4Enc, source)
	if errInputStream != nil {
		fmt.Fprintf(os.Stderr, "ERROR: opening crypto stream for blob %s; %v\n", blob.Hash, errOpening)
		return
	}
	defer is.Close()

	hash, errHashing := util.DataHash(is, modl.Key4Hashes)
	if errHashing != nil {
		fmt.Fprintf(os.Stderr, "ERROR: hashing blob %s; %v\n", blob.Hash, errHashing)
	} else if hash != blob.Hash {
		fmt.Fprintf(os.Stderr, "ERROR: bad checksum for blob %s; %v\n", blob.Hash, errHashing)
	}
}

func ckBlobInAgglo(modl *model.Model, bkpDir string, blob model.Blob) {
	agglo := *blob.AggloRef
	source, errOpening := os.Open(path.Join(bkpDir, agglo.AggloID[1:2], agglo.AggloID))
	if errOpening != nil {
		fmt.Fprintf(os.Stderr, "ERROR: opening blob %s; %v\n", blob.Hash, errOpening)
		return
	}
	defer source.Close()

	ais, errSlicingAgglo := agglos.NewAIS(agglo.Offset, blob.BlobSize, source)
	if errSlicingAgglo != nil {
		fmt.Fprintf(os.Stderr, "ERROR: opening agglo slice stream %s; %v\n", blob.Hash, errOpening)
		return
	}

	is, errInputStream := streams.NewIS(modl.Key4Enc, ais)
	if errInputStream != nil {
		fmt.Fprintf(os.Stderr, "ERROR: opening crypto stream for blob %s; %v\n", blob.Hash, errOpening)
		return
	}
	defer is.Close()

	hash, errHashing := util.DataHash(is, modl.Key4Hashes)
	if errHashing != nil {
		fmt.Fprintf(os.Stderr, "ERROR: hashing blob %s; %v\n", blob.Hash, errHashing)
	} else if hash != blob.Hash {
		fmt.Fprintf(os.Stderr, "ERROR: bad checksum for blob %s; %v\n", blob.Hash, errHashing)
	}
}
