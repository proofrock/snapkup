package agglo

import (
	"github.com/proofrock/snapkup/model"
	"github.com/proofrock/snapkup/util"
	"github.com/proofrock/snapkup/util/agglos"
)

func Unpack(bkpDir string) func(modl *model.Model) error {
	return func(modl *model.Model) error {
		errSplitting := agglos.SplitAll(modl, bkpDir)
		if errSplitting != nil {
			return errSplitting
		}

		util.PrintlnOut("All ok.")

		return nil
	}
}
