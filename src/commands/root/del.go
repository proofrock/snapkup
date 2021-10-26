package root

import (
	"fmt"

	"github.com/proofrock/snapkup/model"
)

func Del(toDel string) func(modl *model.Model) error {
	return func(modl *model.Model) error {
		var found = -1
		for i, root := range modl.Roots {
			if root.Path == toDel {
				found = i
			}
		}

		if found == -1 {
			return fmt.Errorf("Root not found in pool (%s)", toDel)
		}

		modl.Roots = append(modl.Roots[:found], modl.Roots[found+1:]...)

		println("Root correctly deleted (", toDel, ")")

		return nil
	}
}
