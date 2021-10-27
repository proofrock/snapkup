package init

import (
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/proofrock/snapkup/model"
	"github.com/proofrock/snapkup/util"
)

func Init(key []byte, bkpDir string) error {
	if isEmpty, errCheckingEmpty := util.IsEmpty(bkpDir); errCheckingEmpty != nil {
		return errCheckingEmpty
	} else if !isEmpty {
		return fmt.Errorf("Backup dir is not empty (%s)", bkpDir)
	}

	data, errNewData := model.NewModel()
	if errNewData != nil {
		return errNewData
	}

	if errSaveModel := model.SaveModel(key, bkpDir, *data); errSaveModel != nil {
		return errSaveModel
	}

	hex := []rune("0123456789abcdef")
	for i := 0; i < len(hex); i++ {
		if errCreatingDir := os.Mkdir(path.Join(bkpDir, string(hex[i])), fs.FileMode(0700)); errCreatingDir != nil {
			return errCreatingDir
		}
	}

	fmt.Printf("Backup directory correctly initialized in %s\n", bkpDir)

	return nil
}
