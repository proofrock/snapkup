package agglo

import (
	"github.com/proofrock/snapkup/model"
	"github.com/proofrock/snapkup/util/agglos"
)

func Unpack(bkpDir string) func(modl *model.Model) error {
	return func(modl *model.Model) error {
		errSplitting := agglos.SplitAll(modl, bkpDir)
		if errSplitting != nil {
			return errSplitting
		}

		println("All ok.")

		return nil
	}
}
