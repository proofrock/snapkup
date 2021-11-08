package main

import (
	"bufio"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/proofrock/snapkup/commands/agglo"

	"gopkg.in/alecthomas/kingpin.v2"

	initcmd "github.com/proofrock/snapkup/commands/init"
	"github.com/proofrock/snapkup/commands/root"
	"github.com/proofrock/snapkup/commands/snap"
	"github.com/proofrock/snapkup/model"

	"github.com/proofrock/snapkup/util"
)

const version = "v0.3.2"

var (
	relBkpDir       = kingpin.Flag("backup-dir", "The directory to store backups into.").Required().Short('d').ExistingDir()
	profileArg      = kingpin.Flag("profile", "The profile for which to get the password from the credentials file.").Short('p').String()
	noPwdPromptFlag = kingpin.Flag("no-pwd-prompt", "Won't fallback to prompt for password if other methods fail.").Bool()

	initCmd = kingpin.Command("init", "Initializes an empty backups directory.")

	rootCmd = kingpin.Command("root", "Commands related to the root(s)").Alias("r")

	addRootCmd   = rootCmd.Command("add", "Adds a new root to the pool.").Alias("a")
	relRootToAdd = addRootCmd.Arg("root", "The new root to add.").Required().ExistingFileOrDir()

	listRootsCmd = rootCmd.Command("list", "Lists the roots currently in the pool").Alias("ls")

	delRootCmd = rootCmd.Command("del", "Removes a root from the pool.").Alias("rm")
	rootToDel  = delRootCmd.Arg("root", "The root to remove.").Required().String()

	snpCmd = kingpin.Command("snap", "Commands related to the snap(s)").Alias("s")

	snapCmd        = snpCmd.Command("do", "Takes a new snapshot of the roots.")
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

	checkSnapCmd = snpCmd.Command("check", "Checks the on-disk structures for problems.").Alias("ck")

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

func exec(bkpDir string, save bool, block func(modl *model.Model) error) error {
	pwd, errGettingPwd := getPwd(false)
	if errGettingPwd != nil {
		return errGettingPwd
	}

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

func app() (errApp error) {
	kingpin.Version(util.Banner(version))

	cliResult := kingpin.Parse()

	if bkpDir, errAbsolutizing := filepath.Abs(*relBkpDir); errAbsolutizing != nil {
		errApp = errAbsolutizing
	} else {
		switch cliResult {

		case initCmd.FullCommand():

			if pwd, errGettingPwd := getPwd(true); errGettingPwd != nil {
				errApp = errGettingPwd
			} else {
				errApp = initcmd.Init(pwd, bkpDir)
			}

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
			errApp = exec(bkpDir, true, snap.Delete(bkpDir, *snapToDel))

		case restoreCmd.FullCommand():
			if dirToRestore, errAbsolutizing := filepath.Abs(*relDirToRestore); errAbsolutizing != nil {
				errApp = errAbsolutizing
			} else {
				errApp = exec(bkpDir, false, snap.Restore(bkpDir, *snapToRestore, dirToRestore, *restorePrefixPath))
			}

		case infoSnapCmd.FullCommand():
			errApp = exec(bkpDir, false, snap.Info(*snapToInfo))

		case listSnapCmd.FullCommand():
			errApp = exec(bkpDir, false, snap.FileList(*snapToList))

		case labelSnapCmd.FullCommand():
			errApp = exec(bkpDir, true, snap.Label(*snapToLabel, *labelSnapLabel))

		case checkSnapCmd.FullCommand():
			errApp = exec(bkpDir, false, snap.Check(bkpDir))

		case aggloCalcCmd.FullCommand():
			errApp = exec(bkpDir, false, agglo.Calc(*acThreshold*util.Mega, *acTarget*util.Mega))

		case aggloDoCmd.FullCommand():
			errApp = exec(bkpDir, true, agglo.Do(bkpDir, *adThreshold*util.Mega, *adTarget*util.Mega))

		case aggloUnpackCmd.FullCommand():
			errApp = exec(bkpDir, true, agglo.Unpack(bkpDir))
		}
	}

	return errApp
}

func getPwd(first bool) (string, error) {
	if *profileArg != "" {
		// try loading ~/.snapkup-creds
		homeDir, errGettingHomeDir := os.UserHomeDir()
		if errGettingHomeDir != nil {
			return "", errGettingHomeDir
		}
		path := path.Join(homeDir, ".snapkup-creds")
		stat, errStating := os.Stat(path)
		if errStating != nil {
			return "", errors.New("credentials file (~/.snapkup-creds) not found")
		}
		if runtime.GOOS != "windows" && stat.Mode().Perm() != 0600 {
			return "", errors.New("credentials file shouldn't be group- or world- readable (0600 permissions)")
		}
		creds, errReading := os.Open(path)
		if errReading != nil {
			return "", errors.New(fmt.Sprintf("error opening credentials file: %v", errReading))
		}
		defer creds.Close()
		scanner := bufio.NewScanner(creds)
		for scanner.Scan() {
			row := scanner.Text()
			if strings.HasPrefix(row, "#") || strings.TrimSpace(row) == "" {
				continue
			}
			index := strings.Index(row, ":")
			if index < 0 {
				return "", errors.New(fmt.Sprintf("malformed credentials row: %s", row))
			}
			if row[:index] == *profileArg {
				return row[index+1:], nil
			}
		}

		if errScanning := scanner.Err(); errScanning != nil {
			return "", errScanning
		}

		return "", errors.New(fmt.Sprintf("password not found for profile %s", *profileArg))
	}

	if pwd, present := os.LookupEnv("SNAPKUP_PASSWORD"); !present {
		if *noPwdPromptFlag {
			return "", errors.New("Password not provided as env var, nor via --profile argument. Aborting.")
		}
		println("Password not provided as env var, nor via --profile argument.")
		print("Please provide a password")
		if first {
			println(" for init.")
			pwd1 := util.GetPassword("Password: ")
			pwd2 := util.GetPassword("Repeat: ")
			if pwd1 != pwd2 {
				return "", errors.New("passwords do not match")
			}
			return pwd1, nil
		} else {
			println(".")
			return util.GetPassword("Password: "), nil
		}
	} else {
		return pwd, nil
	}
}

func main() {
	if errApp := app(); errApp != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", errApp)
		os.Exit(1)
	}
}
