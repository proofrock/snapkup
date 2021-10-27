package snap

import "github.com/proofrock/snapkup/model"

func findSnap(modl *model.Model, snap int) int {
	for i, snp := range modl.Snaps {
		if snp.Id == snap {
			return i
		}
	}

	return -1
}
