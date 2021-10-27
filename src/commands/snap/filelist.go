package snap

import (
	"fmt"
	"sort"

	"github.com/proofrock/snapkup/model"
)

func FileList(snap int) func(modl *model.Model) error {
	return func(modl *model.Model) error {
		if findSnap(modl, snap) == -1 {
			return fmt.Errorf("Snap %d not found in pool", snap)
		}

		var list []string
		for _, item := range modl.Items {
			if item.Snap == snap {
				list = append(list, item.Path)
			}
		}

		sort.Strings(list)

		for _, path := range list {
			println(path)
		}

		return nil
	}
}
