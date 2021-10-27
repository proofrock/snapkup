package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"

	initcmd "github.com/proofrock/snapkup/commands/init"
	"github.com/proofrock/snapkup/commands/root"
	"github.com/proofrock/snapkup/commands/snap"
	"github.com/proofrock/snapkup/model"

	"github.com/proofrock/snapkup/util"
)

const version = "v0.2.0"

var (
	relBkpDir = kingpin.Flag("backup-dir", "The directory to store backups into.").Required().Short('d').ExistingDir()

	initCmd = kingpin.Command("init", "Initializes an empty backups directory.")

	rootCmd = kingpin.Command("root", "Commands related to the root(s)").Alias("r")

	addRootCmd   = rootCmd.Command("add", "Adds a new root to the pool.").Alias("a")
	relRootToAdd = addRootCmd.Arg("root", "The new root to add.").Required().ExistingFileOrDir()

	listRootsCmd = rootCmd.Command("list", "Lists the roots currently in the pool").Alias("ls")

	delRootCmd = rootCmd.Command("del", "Removes a root from the pool.").Alias("rm")
	rootToDel  = delRootCmd.Arg("root", "The root to remove.").Required().String()

	snpCmd = kingpin.Command("snap", "Commands related to the snap(s)").Alias("s")

	snapCmd        = snpCmd.Command("take", "Takes a new snapshot of the roots.").Alias("do")
	snapNoCompress = snapCmd.Flag("no-compress", "Doesn't compress the stored files.").Bool()
	snapLabel      = snapCmd.Flag("label", "Label for this snap.").Short('l').Default("").String()

	listSnapsCmd = snpCmd.Command("list", "Lists the snaps currently in the pool").Alias("ls")

	delSnapCmd = snpCmd.Command("del", "Removes a snap from the pool.").Alias("rm")
	snapToDel  = delSnapCmd.Arg("snap", "The snap to remove.").Required().Int()

	restoreCmd        = snpCmd.Command("restore", "Restores a snap.").Alias("res")
	snapToRestore     = restoreCmd.Arg("snap", "The snap to restore.").Required().Int()
	relDirToRestore   = restoreCmd.Arg("restore-dir", "The dir to restore into. Must exist and be empty.").Required().ExistingDir()
	restorePrefixPath = restoreCmd.Flag("prefix-path", "Only the files whose path starts with this prefix are considered.").String()

	infoSnapCmd = snpCmd.Command("info", "Gives relevant information on a snap or on all snaps.")
	snapToInfo  = infoSnapCmd.Arg("snap", "The snap to give info about.").Int()

	listSnapCmd = snpCmd.Command("filelist", "Prints the list of files for a snap.").Alias("fl")
	snapToList  = listSnapCmd.Arg("snap", "The snap to list files for.").Required().Int()

	labelSnapCmd   = snpCmd.Command("label", "Sets or changes the label of a snap.").Alias("lbl")
	snapToLabel    = labelSnapCmd.Arg("snap", "The snap to label.").Required().Int()
	labelSnapLabel = labelSnapCmd.Arg("label", "The label.").Required().String()
)

func exec(bkpDir string, save bool, block func(modl *model.Model) error) error {
	modl, errLoadingModel := model.LoadModel(util.FakeKey, bkpDir)
	if errLoadingModel != nil {
		return errLoadingModel
	}
	if errExecutingPayload := block(modl); errExecutingPayload != nil {
		return errExecutingPayload
	}
	if save {
		if errSavingModel := model.SaveModel(util.FakeKey, bkpDir, *modl); errSavingModel != nil {
			return errSavingModel
		}
	}
	return nil
}

func app() (errApp error) {
	kingpin.Version(util.Banner(version))

	cliResult := kingpin.Parse()

	if bkpDir, errAbsolutizing := filepath.Abs(*relBkpDir); errAbsolutizing != nil {
		errApp = errAbsolutizing
	} else {
		switch cliResult {

		case initCmd.FullCommand():
			errApp = initcmd.Init(util.FakeKey, bkpDir)

		case addRootCmd.FullCommand():
			if rootToAdd, errAbsolutizing := filepath.Abs(*relRootToAdd); errAbsolutizing != nil {
				errApp = errAbsolutizing
			} else {
				errApp = exec(bkpDir, true, root.Add(rootToAdd))
			}

		case listRootsCmd.FullCommand():
			errApp = exec(bkpDir, false, root.List())

		case delRootCmd.FullCommand():
			errApp = exec(bkpDir, true, root.Delete(*rootToDel))

		case snapCmd.FullCommand():
			errApp = exec(bkpDir, true, snap.Do(bkpDir, *snapNoCompress, *snapLabel))

		case listSnapsCmd.FullCommand():
			errApp = exec(bkpDir, false, snap.List())

		case delSnapCmd.FullCommand():
			errApp = exec(bkpDir, true, snap.Delete(*snapToDel))

		case restoreCmd.FullCommand():
			if dirToRestore, errAbsolutizing := filepath.Abs(*relDirToRestore); errAbsolutizing != nil {
				errApp = errAbsolutizing
			} else {
				errApp = exec(bkpDir, false, snap.Restore(bkpDir, *snapToRestore, dirToRestore, restorePrefixPath))
			}

		case infoSnapCmd.FullCommand():
			errApp = exec(bkpDir, false, snap.Info(snapToInfo))

		case listSnapCmd.FullCommand():
			errApp = exec(bkpDir, false, snap.FileList(*snapToList))

		case labelSnapCmd.FullCommand():
			errApp = exec(bkpDir, true, snap.Label(*snapToLabel, *labelSnapLabel))
		}
	}

	return errApp
}

func main() {
	if errApp := app(); errApp != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", errApp)
		os.Exit(1)
	}
}
