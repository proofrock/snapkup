package root

import (
	"github.com/proofrock/snapkup/model"
	"github.com/proofrock/snapkup/util"
)

func List() func(modl *model.Model) error {
	return func(modl *model.Model) error {
		for _, root := range modl.Roots {
			util.PrintlnOut(root)
		}

		return nil
	}
}
