package root

import "github.com/proofrock/snapkup/model"

func List() func(modl *model.Model) error {
	return func(modl *model.Model) error {
		for _, root := range modl.Roots {
			println(root.Path)
		}

		return nil
	}
}
