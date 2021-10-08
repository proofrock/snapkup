package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"

	addroot "github.com/proofrock/snapkup/commands/add_root"
	delroot "github.com/proofrock/snapkup/commands/del_root"
	delsnaps "github.com/proofrock/snapkup/commands/del_snap"
	initcmd "github.com/proofrock/snapkup/commands/init"
	listroots "github.com/proofrock/snapkup/commands/list_roots"
	listsnaps "github.com/proofrock/snapkup/commands/list_snaps"
	"github.com/proofrock/snapkup/commands/restore"
	snap "github.com/proofrock/snapkup/commands/snap"

	"github.com/proofrock/snapkup/util"
)

const version = "v0.0.1A1"

var (
	relBkpDir = kingpin.Flag("backup-dir", "The directory to store backups into.").Required().Short('d').ExistingDir()

	initCmd = kingpin.Command("init", "Initializes an empty backups directory.")

	addRootCmd   = kingpin.Command("add-root", "Adds a new root to the pool.")
	relRootToAdd = addRootCmd.Arg("root", "The new root to add.").Required().ExistingDir()

	listRootsCmd = kingpin.Command("list-roots", "Lists the roots currently in the pool")

	delRootCmd = kingpin.Command("del-root", "Removes a root from the pool.")
	rootToDel  = delRootCmd.Arg("root", "The root to remove.").Required().String()

	snapCmd = kingpin.Command("snap", "Takes a new snapshot of the roots.")

	listSnapsCmd = kingpin.Command("list-snaps", "Lists the snaps currently in the pool")

	delSnapCmd = kingpin.Command("del-snap", "Removes a snap from the pool.")
	snapToDel  = delSnapCmd.Arg("snap", "The snap to remove.").Required().Int()

	restoreCmd      = kingpin.Command("restore", "Restores a snap.")
	snapToRestore   = restoreCmd.Arg("snap", "The snap to restore.").Required().Int()
	relDirToRestore = restoreCmd.Arg("restore-dir", "The dir to restore into. Must exist and be empty.").Required().ExistingDir()
)

func main() {
	kingpin.Version(util.Banner(version))

	cliResult := kingpin.Parse()

	var errApp error = nil

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
			errApp = snap.Snap(bkpDir)

		case listSnapsCmd.FullCommand():
			errApp = listsnaps.ListSnaps(bkpDir)

		case delSnapCmd.FullCommand():
			errApp = delsnaps.DelSnap(bkpDir, *snapToDel)

		case restoreCmd.FullCommand():
			if dirToRestore, errAbsolutizing := filepath.Abs(*relDirToRestore); errAbsolutizing != nil {
				errApp = errAbsolutizing
			} else {
				errApp = restore.Restore(bkpDir, *snapToRestore, dirToRestore)
			}
		}
	}

	if errApp != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", errApp)
	}
}
