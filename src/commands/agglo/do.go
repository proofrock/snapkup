package agglo

import (
	"github.com/proofrock/snapkup/model"
	"github.com/proofrock/snapkup/util"
	"github.com/proofrock/snapkup/util/agglos"
)

func Do(bkpDir string, threshold, target int) func(modl *model.Model) error {
	return func(modl *model.Model) error {
		aggloss, blobs, errPlanning := agglos.Plan(modl, int64(threshold), int64(target))
		if errPlanning != nil {
			return errPlanning
		}

		util.PrintlnfOut("%d files will be merged, resulting in %d agglo files.", len(blobs), len(aggloss))
		util.PrintlnOut("Performing merge...")

		errApplying := agglos.Apply(modl, bkpDir, aggloss, blobs)
		if errApplying != nil {
			return errApplying
		}

		util.PrintlnOut("All done.")

		return nil
	}
}
