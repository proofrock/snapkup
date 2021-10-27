package storage

import (
	"fmt"
	"os"
	"path"

	"github.com/proofrock/snapkup/model"
)

func GarbageCollect(bkpDir string) func(modl *model.Model) error {
	return func(modl *model.Model) error {
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
