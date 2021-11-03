package snap

import (
	"fmt"
	"time"

	"github.com/proofrock/snapkup/model"
)

func List() func(modl *model.Model) error {
	return func(modl *model.Model) error {
		for _, snap := range modl.Snaps {
			ts := time.UnixMilli(snap.Timestamp).Local().Format("2 Jan 2006, 15:04:05 (MST)")
			fmt.Printf("Snap %d:\t%s\t%s\n", snap.Id, ts, snap.Label)
		}

		return nil
	}
}
