package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"

	addroot "github.com/proofrock/snapkup/commands/add_root"
	delroot "github.com/proofrock/snapkup/commands/del_root"
	delsnaps "github.com/proofrock/snapkup/commands/del_snap"
	"github.com/proofrock/snapkup/commands/info_snap"
	initcmd "github.com/proofrock/snapkup/commands/init"
	labelsnap "github.com/proofrock/snapkup/commands/label_snap"
	listroots "github.com/proofrock/snapkup/commands/list_roots"
	"github.com/proofrock/snapkup/commands/list_snap"
	listsnaps "github.com/proofrock/snapkup/commands/list_snaps"
	"github.com/proofrock/snapkup/commands/restore"
	snap "github.com/proofrock/snapkup/commands/snap"

	"github.com/proofrock/snapkup/util"
)

const version = "v0.1.0"

var (
	relBkpDir = kingpin.Flag("backup-dir", "The directory to store backups into.").Required().Short('d').ExistingDir()

	initCmd = kingpin.Command("init", "Initializes an empty backups directory.")

	addRootCmd   = kingpin.Command("add-root", "Adds a new root to the pool.")
	relRootToAdd = addRootCmd.Arg("root", "The new root to add.").Required().ExistingFileOrDir()

	listRootsCmd = kingpin.Command("list-roots", "Lists the roots currently in the pool")

	delRootCmd = kingpin.Command("del-root", "Removes a root from the pool.")
	rootToDel  = delRootCmd.Arg("root", "The root to remove.").Required().String()

	snapCmd      = kingpin.Command("snap", "Takes a new snapshot of the roots.")
	snapCompress = snapCmd.Flag("compress", "Compresses the stored files.").Short('z').Bool()
	snapLabel    = snapCmd.Flag("label", "Label for this snap.").Short('l').Default("").String()

	listSnapsCmd = kingpin.Command("list-snaps", "Lists the snaps currently in the pool")

	delSnapCmd = kingpin.Command("del-snap", "Removes a snap from the pool.")
	snapToDel  = delSnapCmd.Arg("snap", "The snap to remove.").Required().Int()

	restoreCmd        = kingpin.Command("restore", "Restores a snap.")
	snapToRestore     = restoreCmd.Arg("snap", "The snap to restore.").Required().Int()
	relDirToRestore   = restoreCmd.Arg("restore-dir", "The dir to restore into. Must exist and be empty.").Required().ExistingDir()
	restorePrefixPath = restoreCmd.Flag("prefix-path", "Only the files whose path starts with this prefix are considered.").String()

	infoSnapCmd = kingpin.Command("info-snap", "Gives relevant information on a snap.")
	snapToInfo  = infoSnapCmd.Arg("snap", "The snap to give info about.").Required().Int()

	listSnapCmd = kingpin.Command("list-snap", "Prints the list of files for a snap.")
	snapToList  = listSnapCmd.Arg("snap", "The snap to list files for.").Required().Int()

	labelSnapCmd   = kingpin.Command("label-snap", "Sets or changes the label of a snap.")
	snapToLabel    = labelSnapCmd.Arg("snap", "The snap to label.").Required().Int()
	labelSnapLabel = labelSnapCmd.Arg("label", "The label.").Required().String()
)

func app() (errApp error) {
	kingpin.Version(util.Banner(version))

	cliResult := kingpin.Parse()

	if bkpDir, errAbsolutizing := filepath.Abs(*relBkpDir); errAbsolutizing != nil {
		errApp = errAbsolutizing
	} else {
		switch cliResult {

		case initCmd.FullCommand():
			errApp = initcmd.Init(bkpDir)

		case addRootCmd.FullCommand():
			if rootToAdd, errAbsolutizing := filepath.Abs(*relRootToAdd); errAbsolutizing != nil {
				errApp = errAbsolutizing
			} else {
				errApp = addroot.AddRoot(bkpDir, rootToAdd)
			}

		case listRootsCmd.FullCommand():
			errApp = listroots.ListRoots(bkpDir)

		case delRootCmd.FullCommand():
			errApp = delroot.DelRoot(bkpDir, *rootToDel)

		case snapCmd.FullCommand():
			errApp = snap.Snap(bkpDir, *snapCompress, *snapLabel)

		case listSnapsCmd.FullCommand():
			errApp = listsnaps.ListSnaps(bkpDir)

		case delSnapCmd.FullCommand():
			errApp = delsnaps.DelSnap(bkpDir, *snapToDel)

		case restoreCmd.FullCommand():
			if dirToRestore, errAbsolutizing := filepath.Abs(*relDirToRestore); errAbsolutizing != nil {
				errApp = errAbsolutizing
			} else {
				errApp = restore.Restore(bkpDir, *snapToRestore, dirToRestore, restorePrefixPath)
			}

		case infoSnapCmd.FullCommand():
			errApp = info_snap.InfoSnap(bkpDir, *snapToInfo)

		case listSnapCmd.FullCommand():
			errApp = list_snap.ListSnap(bkpDir, *snapToList)

		case labelSnapCmd.FullCommand():
			errApp = labelsnap.LabelSnap(bkpDir, *snapToLabel, *labelSnapLabel)
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
