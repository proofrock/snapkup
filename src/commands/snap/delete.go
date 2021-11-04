package snap

import (
	"fmt"
	"os"

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
			if blob.AggloRef == nil {
				blobsToDel[blob.Hash] = true
			}
		}

		for _, item := range modl.Items {
			delete(blobsToDel, item.Hash)
		}

		fmt.Printf("Deleting %d blobs...\n", len(blobsToDel))
		for hash := range blobsToDel {
			pathToDel := model.HashToPath(bkpDir, hash)
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
