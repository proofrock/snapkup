package init

import (
	"fmt"
	"io/fs"
	"os"
	"path"

	"github.com/proofrock/snapkup/model"
	"github.com/proofrock/snapkup/util"
)

func Init(pwd string, bkpDir string) error {
	// FIXME it fails for .DS_Store files in MacOS
	if isEmpty, errCheckingEmpty := util.IsEmpty(bkpDir); errCheckingEmpty != nil {
		return errCheckingEmpty
	} else if !isEmpty {
		return fmt.Errorf("backup dir is not empty (%s)", bkpDir)
	}

	data, errNewData := model.NewModel()
	if errNewData != nil {
		return errNewData
	}

	if errSaveModel := model.SaveModel(pwd, bkpDir, *data); errSaveModel != nil {
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
