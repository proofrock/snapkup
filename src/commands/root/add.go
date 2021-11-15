package root

import (
	"fmt"

	"github.com/proofrock/snapkup/model"
	"github.com/proofrock/snapkup/util"
)

func Add(toAdd string) func(modl *model.Model) error {
	return func(modl *model.Model) error {
		for _, root := range modl.Roots {
			if root == toAdd {
				return fmt.Errorf("root already present (%s)", toAdd)
			}
		}

		modl.Roots = append(modl.Roots, toAdd)

		util.PrintlnfOut("Root correctly added (%s)", toAdd)

		return nil
	}
}
