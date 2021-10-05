package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/proofrock/snapkup/commands"
	"github.com/proofrock/snapkup/util"
	"gopkg.in/alecthomas/kingpin.v2"
)

const version = "v0.0.1A1"

var (
	app     = kingpin.New("snapkup", "Incremental backups for the masses.")
	_bkpDir = kingpin.Flag("backup-dir", "The directory to store backups into.").Required().Short('d').ExistingDir()
	quiet   = kingpin.Flag("quiet", "Don't print the banner.").Short('q').Bool()

	initCmd = kingpin.Command("init", "Initializes an empty backups directory.")

	addRootCmd = kingpin.Command("add-root", "Adds a new root to the pool.")
	rootToAdd  = addRootCmd.Arg("root", "The new root to add.").Required().ExistingDir()

	listRootsCmd = kingpin.Command("list-roots", "Lists the roots currently in the pool")

	delRootCmd = kingpin.Command("del-root", "Removes a root from the pool.")
	rootToDel  = delRootCmd.Arg("root", "The root to remove.").Required().String()
)

func main() {
	kingpin.Version(version)

	cliResult := kingpin.Parse()

	if !*quiet {
		util.PrintBanner(version)
	}

	var err error = nil

	if bkpDir, errAbsolutizing := filepath.Abs(*_bkpDir); errAbsolutizing != nil {
		err = errAbsolutizing
	} else {
		switch cliResult {
		case initCmd.FullCommand():
			err = commands.Init(bkpDir)
		case addRootCmd.FullCommand():
			err = commands.AddRoot(bkpDir, rootToAdd)
		case listRootsCmd.FullCommand():
			err = commands.ListRoots(bkpDir)
		case delRootCmd.FullCommand():
			err = commands.DelRoot(bkpDir, rootToDel)
		}
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	}
}
