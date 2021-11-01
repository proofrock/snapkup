package snap

import (
	"fmt"

	"github.com/proofrock/snapkup/model"
)

func Label(snap int, label string) func(modl *model.Model) error {
	return func(modl *model.Model) error {
		for i := 0; i < len(modl.Snaps); i++ {
			if modl.Snaps[i].Id == snap {
				modl.Snaps[i].Label = label
				println("Ok.")
				return nil
			}
		}

		return fmt.Errorf("snap %d not found in pool", snap)
	}
}
