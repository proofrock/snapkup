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
	snap "github.com/proofrock/snapkup/commands/snap"

	"github.com/proofrock/snapkup/util"
)

const version = "v0.0.1A1"

var (
	_bkpDir = kingpin.Flag("backup-dir", "The directory to store backups into.").Required().Short('d').ExistingDir()

	initCmd = kingpin.Command("init", "Initializes an empty backups directory.")

	addRootCmd = kingpin.Command("add-root", "Adds a new root to the pool.")
	rootToAdd  = addRootCmd.Arg("root", "The new root to add.").Required().ExistingDir()

	listRootsCmd = kingpin.Command("list-roots", "Lists the roots currently in the pool")

	delRootCmd = kingpin.Command("del-root", "Removes a root from the pool.")
	rootToDel  = delRootCmd.Arg("root", "The root to remove.").Required().String()

	snapCmd = kingpin.Command("snap", "Takes a new snapshot of the roots.")

	listSnapsCmd = kingpin.Command("list-snaps", "Lists the snaps currently in the pool")

	delSnapCmd = kingpin.Command("del-snap", "Removes a snap from the pool.")
	snapToDel  = delSnapCmd.Arg("snap", "The snap to remove.").Required().Int()
)

func main() {
	kingpin.Version(util.Banner(version))

	cliResult := kingpin.Parse()

	var err error = nil

	if bkpDir, errAbsolutizing := filepath.Abs(*_bkpDir); errAbsolutizing != nil {
		err = errAbsolutizing
	} else {
		switch cliResult {
		case initCmd.FullCommand():
			err = initcmd.Init(bkpDir)
		case addRootCmd.FullCommand():
			err = addroot.AddRoot(bkpDir, rootToAdd)
		case listRootsCmd.FullCommand():
			err = listroots.ListRoots(bkpDir)
		case delRootCmd.FullCommand():
			err = delroot.DelRoot(bkpDir, rootToDel)
		case snapCmd.FullCommand():
			err = snap.Snap(bkpDir)
		case listSnapsCmd.FullCommand():
			err = listsnaps.ListSnaps(bkpDir)
		case delSnapCmd.FullCommand():
			err = delsnaps.DelSnap(bkpDir, snapToDel)
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	}
}
