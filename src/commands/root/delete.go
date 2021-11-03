package root

import (
	"fmt"

	"github.com/proofrock/snapkup/model"
)

func Delete(toDel string) func(modl *model.Model) error {
	return func(modl *model.Model) error {
		var found = -1
		for i, root := range modl.Roots {
			if root == toDel {
				found = i
				break
			}
		}

		if found == -1 {
			return fmt.Errorf("root not found in pool (%s)", toDel)
		}

		modl.Roots = append(modl.Roots[:found], modl.Roots[found+1:]...)

		fmt.Printf("Root correctly deleted (%s)\n", toDel)

		return nil
	}
}
