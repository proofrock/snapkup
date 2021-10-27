package snap

import (
	"sort"

	"github.com/proofrock/snapkup/model"
)

func FileList(snap int) func(modl *model.Model) error {
	return func(modl *model.Model) error {
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
