package root

import (
	"fmt"

	"github.com/proofrock/snapkup/model"
)

func Add(toAdd string) func(modl *model.Model) error {
	return func(modl *model.Model) error {
		for _, root := range modl.Roots {
			if root == toAdd {
				return fmt.Errorf("root already present (%s)", toAdd)
			}
		}

		modl.Roots = append(modl.Roots, toAdd)

		fmt.Printf("Root correctly added (%s)\n", toAdd)

		return nil
	}
}
