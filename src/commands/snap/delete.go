package snap

import (
	"fmt"
	"os"

	"github.com/proofrock/snapkup/model"
	"github.com/proofrock/snapkup/util"
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

		util.PrintlnfOut("Snap %d correctly deleted. Removing dangling files...", toDel)

		blobsToDel := make(map[string]bool)
		for _, blob := range modl.Blobs {
			if blob.AggloRef == nil {
				blobsToDel[blob.Hash] = true
			}
		}

		for _, item := range modl.Items {
			delete(blobsToDel, item.Hash)
		}

		util.PrintlnfOut("Deleting %d blobs...", len(blobsToDel))
		for hash := range blobsToDel {
			pathToDel := model.HashToPath(bkpDir, hash)
			if errDeleting := os.Remove(pathToDel); errDeleting != nil {
				util.PrintlnfErr("ERROR: deleting file %s; %v", hash, errDeleting)
			}
		}

		var nuBlobs []model.Blob
		for _, blob := range modl.Blobs {
			if _, isThere := blobsToDel[blob.Hash]; !isThere {
				nuBlobs = append(nuBlobs, blob)
			}
		}
		modl.Blobs = nuBlobs
		util.PrintlnfOut("%d blobs remaining.", len(nuBlobs))

		util.PrintlnOut("All done.")

		return nil
	}
}
