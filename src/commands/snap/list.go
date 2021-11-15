package snap

import (
	"time"

	"github.com/proofrock/snapkup/model"
	"github.com/proofrock/snapkup/util"
)

func List() func(modl *model.Model) error {
	return func(modl *model.Model) error {
		for _, snap := range modl.Snaps {
			ts := time.UnixMilli(snap.Timestamp).Local().Format("2 Jan 2006, 15:04:05 (MST)")
			util.PrintlnfOut("Snap %d:\t%s\t%s", snap.Id, ts, snap.Label)
		}

		return nil
	}
}
