package root

import (
	"fmt"

	"github.com/proofrock/snapkup/model"
)

func Add(toAdd string) func(modl *model.Model) error {
	return func(modl *model.Model) error {
		for _, root := range modl.Roots {
			if root.Path == toAdd {
				return fmt.Errorf("Root already present (%s)", toAdd)
			}
		}

		modl.Roots = append(modl.Roots, model.Root{Path: toAdd})

		println("Root correctly added (", toAdd, ")")

		return nil
	}
}
