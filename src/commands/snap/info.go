package snap

import (
	"fmt"

	"github.com/proofrock/snapkup/model"
)

const suffixes = "KMGTPE"
const unit int64 = 1024

func fmtBytes(b int64) string {
	if b < unit {
		return fmt.Sprintf("%d b", b)
	}
	div, exp := unit, 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cb",
		float64(b)/float64(div), suffixes[exp])
}

type info struct {
	Files      int
	Dirs       int
	Size       int64
	StoredSize int64
	TotStored  int64
	Referenced int64
}

func Info(snap int) func(modl *model.Model) error {
	return func(modl *model.Model) error {
		if snap > -1 {
			if findSnap(modl, snap) == -1 {
				return fmt.Errorf("snap %d not found in pool", snap)
			}
		}

		var nfo info

		blobs := make(map[string]model.Blob)
		for _, blob := range modl.Blobs {
			blobs[blob.Hash] = blob
			nfo.TotStored += blob.BlobSize
		}

		alreadyDiscoveredBlobs := make(map[string]bool)
		alreadyDiscoveredBlobsForSnap := make(map[string]bool)
		for _, item := range modl.Items {
			if _, alreadyDiscovered := alreadyDiscoveredBlobs[item.Hash]; !alreadyDiscovered {
				nfo.Referenced += blobs[item.Hash].BlobSize
				alreadyDiscoveredBlobs[item.Hash] = true
			}
			if snap <= -1 || item.Snap == snap {
				if _, alreadyDiscovered := alreadyDiscoveredBlobsForSnap[item.Hash]; !alreadyDiscovered {
					nfo.StoredSize += blobs[item.Hash].BlobSize
					alreadyDiscoveredBlobsForSnap[item.Hash] = true
				}
				nfo.Size += blobs[item.Hash].Size
				if item.IsDir {
					nfo.Dirs++
				} else {
					nfo.Files++
				}
			}
		}

		fmt.Printf("Files                    : %d\n", nfo.Files)
		fmt.Printf("Directories              : %d\n", nfo.Dirs)
		fmt.Printf("Size                     : %s\n", fmtBytes(nfo.Size))
		if snap > -1 {
			fmt.Printf("Stored size              : %s\n", fmtBytes(nfo.StoredSize))
		}
		fmt.Printf("Tot. stored (all snaps)  : %s\n", fmtBytes(nfo.TotStored))
		if snap <= -1 {
			fmt.Printf("Can be garbage collected : %s\n", fmtBytes(nfo.TotStored-nfo.Referenced))
		}

		return nil
	}
}
