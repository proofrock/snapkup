package agglo

import (
	"fmt"
	"github.com/proofrock/snapkup/model"
	"github.com/proofrock/snapkup/util/agglos"
)

func Calc(threshold, target int) func(modl *model.Model) error {
	return func(modl *model.Model) error {
		agglos, blobs, errPlanning := agglos.Plan(modl, int64(threshold), int64(target))
		if errPlanning != nil {
			return errPlanning
		}

		fmt.Printf("%d files will be merged, resulting in %d agglo files.\n", len(blobs), len(agglos))

		return nil
	}
}
