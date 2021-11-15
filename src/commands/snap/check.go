package snap

import (
	"os"

	"github.com/proofrock/snapkup/model"
	"github.com/proofrock/snapkup/util"
	"github.com/proofrock/snapkup/util/agglos"
	"github.com/proofrock/snapkup/util/streams"
)

func Check(bkpDir string) func(modl *model.Model) error {
	return func(modl *model.Model) error {
		allItems := make(map[string]bool)
		if len(modl.Agglos) > 0 {
			util.PrintlnOut("Checking agglos... ")
			for _, agglo := range modl.Agglos {
				ckAgglo(modl, bkpDir, agglo)
				allItems[agglo.ID] = true
			}
			util.PrintlnOut("Done.")
		}

		util.PrintlnOut("Checking blobs...")
		for _, blob := range modl.Blobs {
			if blob.AggloRef == nil { // normal blob
				ckNormalBlob(modl, bkpDir, blob)
			} else { // blob stored in an agglo
				if allItems[(*blob.AggloRef).AggloID] {
					ckBlobInAgglo(modl, bkpDir, blob)
				} else {
					util.PrintlnfErr("ERROR: agglo %s not found for blob %s", (*blob.AggloRef).AggloID, blob.Hash)
				}
			}
			allItems[blob.Hash] = true
		}
		util.PrintlnOut("Done.")

		util.PrintlnOut("Checking filesystem...")
		files, _, _ := walkFSTree([]string{bkpDir}, nil, false)
		for _, file := range files {
			if file.IsDir || file.Name == model.ModelFileName {
				continue
			}
			if !allItems[file.Name] {
				util.PrintlnfErr("WARNING: spurious file in filesystem: %s", file.FullPath)
			}
		}
		util.PrintlnOut("Done.")

		return nil
	}
}

func ckAgglo(modl *model.Model, bkpDir string, agglo model.Agglo) {
	fpath := model.AggloIdToPath(bkpDir, agglo.ID)
	hash, errHashing := util.FileHash(fpath, modl.Key4Hashes)
	if errHashing != nil {
		util.PrintlnfErr("ERROR: hashing agglo %s; %v", agglo.ID, errHashing)
	} else if hash != agglo.Hash {
		util.PrintlnfErr("ERROR: bad checksum for agglo %s; %v", agglo.ID, errHashing)
	}
}

func ckNormalBlob(modl *model.Model, bkpDir string, blob model.Blob) {
	source, errOpening := os.Open(model.HashToPath(bkpDir, blob.Hash))
	if errOpening != nil {
		util.PrintlnfErr("ERROR: opening blob %s; %v", blob.Hash, errOpening)
		return
	}
	defer source.Close()

	is, errInputStream := streams.NewIS(modl.Key4Enc, source)
	if errInputStream != nil {
		util.PrintlnfErr("ERROR: opening crypto stream for blob %s; %v", blob.Hash, errOpening)
		return
	}
	defer is.Close()

	hash, errHashing := util.DataHash(is, modl.Key4Hashes)
	if errHashing != nil {
		util.PrintlnfErr("ERROR: hashing blob %s; %v", blob.Hash, errHashing)
	} else if hash != blob.Hash {
		util.PrintlnfErr("ERROR: bad checksum for blob %s; %v", blob.Hash, errHashing)
	}
}

func ckBlobInAgglo(modl *model.Model, bkpDir string, blob model.Blob) {
	agglo := *blob.AggloRef
	source, errOpening := os.Open(model.AggloIdToPath(bkpDir, agglo.AggloID))
	if errOpening != nil {
		util.PrintlnfErr("ERROR: opening blob %s; %v", blob.Hash, errOpening)
		return
	}
	defer source.Close()

	ais, errSlicingAgglo := agglos.NewAIS(agglo.Offset, blob.BlobSize, source)
	if errSlicingAgglo != nil {
		util.PrintlnfErr("ERROR: opening agglo slice stream %s; %v", blob.Hash, errOpening)
		return
	}

	is, errInputStream := streams.NewIS(modl.Key4Enc, ais)
	if errInputStream != nil {
		util.PrintlnfErr("ERROR: opening crypto stream for blob %s; %v", blob.Hash, errOpening)
		return
	}
	defer is.Close()

	hash, errHashing := util.DataHash(is, modl.Key4Hashes)
	if errHashing != nil {
		util.PrintlnfErr("ERROR: hashing blob %s; %v", blob.Hash, errHashing)
	} else if hash != blob.Hash {
		util.PrintlnfErr("ERROR: bad checksum for blob %s; %v", blob.Hash, errHashing)
	}
}
