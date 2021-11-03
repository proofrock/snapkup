package main

import (
	"fmt"
	"github.com/proofrock/snapkup/commands/agglo"
	"math/rand"
	"os"
	"path/filepath"
	"time"

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
	snapToInfo  = infoSnapCmd.Arg("snap", "The snap to give info about.").Default("-1").Int()

	listSnapCmd = snpCmd.Command("filelist", "Prints the list of files for a snap.").Alias("fl")
	snapToList  = listSnapCmd.Arg("snap", "The snap to list files for.").Required().Int()

	labelSnapCmd   = snpCmd.Command("label", "Sets or changes the label of a snap.").Alias("lbl")
	snapToLabel    = labelSnapCmd.Arg("snap", "The snap to label.").Required().Int()
	labelSnapLabel = labelSnapCmd.Arg("label", "The label.").Required().String()

	aggloCmd = kingpin.Command("agglo", "Commands related to agglo(meration)s of smaller files").Alias("a")

	aggloCalcCmd = aggloCmd.Command("calc", "Calculates how much files can be deleted by agglomerating.").Alias("c")
	acThreshold  = aggloCalcCmd.Arg("threshold", "Files smaller than this size (in Mb) will be merged.").Required().Int()
	acTarget     = aggloCalcCmd.Arg("target", "Target size for the agglomeration files (in Mb).").Required().Int()

	aggloDoCmd  = aggloCmd.Command("do", "Perform agglomerations.")
	adThreshold = aggloDoCmd.Arg("threshold", "Files smaller than this size (in Mb) will be merged.").Required().Int()
	adTarget    = aggloDoCmd.Arg("target", "Target size for the agglomeration files (in Mb).").Required().Int()

	aggloUnpackCmd = aggloCmd.Command("unpack", "Unpacks and removes all the agglomerations.").Alias("x")
)

func init() {
	rand.Seed(time.Now().UnixMilli())
}

func exec(pwd, bkpDir string, save bool, block func(modl *model.Model) error) error {
	modl, errLoadingModel := model.LoadModel(pwd, bkpDir)
	if errLoadingModel != nil {
		return errLoadingModel
	}
	if errExecutingPayload := block(modl); errExecutingPayload != nil {
		return errExecutingPayload
	}
	if save {
		if errSavingModel := model.SaveModel(pwd, bkpDir, *modl); errSavingModel != nil {
			return errSavingModel
		}
	}
	return nil
}

func app(pwd string) (errApp error) {
	kingpin.Version(util.Banner(version))

	cliResult := kingpin.Parse()

	if bkpDir, errAbsolutizing := filepath.Abs(*relBkpDir); errAbsolutizing != nil {
		errApp = errAbsolutizing
	} else {
		switch cliResult {

		case initCmd.FullCommand():
			errApp = initcmd.Init(pwd, bkpDir)

		case addRootCmd.FullCommand():
			if rootToAdd, errAbsolutizing := filepath.Abs(*relRootToAdd); errAbsolutizing != nil {
				errApp = errAbsolutizing
			} else {
				errApp = exec(pwd, bkpDir, true, root.Add(rootToAdd))
			}

		case listRootsCmd.FullCommand():
			errApp = exec(pwd, bkpDir, false, root.List())

		case delRootCmd.FullCommand():
			errApp = exec(pwd, bkpDir, true, root.Delete(*rootToDel))

		case snapCmd.FullCommand():
			errApp = exec(pwd, bkpDir, true, snap.Do(bkpDir, *snapNoCompress, *snapLabel))

		case listSnapsCmd.FullCommand():
			errApp = exec(pwd, bkpDir, false, snap.List())

		case delSnapCmd.FullCommand():
			errApp = exec(pwd, bkpDir, true, snap.Delete(bkpDir, *snapToDel))

		case restoreCmd.FullCommand():
			if dirToRestore, errAbsolutizing := filepath.Abs(*relDirToRestore); errAbsolutizing != nil {
				errApp = errAbsolutizing
			} else {
				errApp = exec(pwd, bkpDir, false, snap.Restore(bkpDir, *snapToRestore, dirToRestore, restorePrefixPath))
			}

		case infoSnapCmd.FullCommand():
			errApp = exec(pwd, bkpDir, false, snap.Info(*snapToInfo))

		case listSnapCmd.FullCommand():
			errApp = exec(pwd, bkpDir, false, snap.FileList(*snapToList))

		case labelSnapCmd.FullCommand():
			errApp = exec(pwd, bkpDir, true, snap.Label(*snapToLabel, *labelSnapLabel))

		case aggloCalcCmd.FullCommand():
			errApp = exec(pwd, bkpDir, false, agglo.Calc(*acThreshold*util.Mega, *acTarget*util.Mega))

		case aggloDoCmd.FullCommand():
			errApp = exec(pwd, bkpDir, true, agglo.Do(bkpDir, *adThreshold*util.Mega, *adTarget*util.Mega))

		case aggloUnpackCmd.FullCommand():
			errApp = exec(pwd, bkpDir, true, agglo.Unpack(bkpDir))
		}
	}

	return errApp
}

func main() {
	pwd := os.Getenv("SNAPKUP_PASSWORD")
	if pwd == "" {
		fmt.Fprint(os.Stderr, "ERROR: password not declared\n")
		os.Exit(1)
	}

	if errApp := app(pwd); errApp != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", errApp)
		os.Exit(1)
	}
}
