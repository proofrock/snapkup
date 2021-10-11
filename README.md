# 🗃️ snapkup

Snapkup is a simple backup tool that takes snapshots of your filesystem (or the parts that you'll decide), storing them efficiently and conveniently.

## Basic workflow

Snapkup's goal is to store efficiently one or more filesystem's situation at given points in time, in a manner that is friendly to e.g. Dropbox sync or removable storage.

- You initialize an empty directory that will store the backups
- You register one or more backup roots, directory or files that will be snapshotted
- You take one or more snapshots. Snapkup lists all the tree for those roots, taking a snapshot of the contents
    - All the files in the roots are deduplicated, and only the files that are different are stored
    - All files that can be compressed are stored as such, using  `zstd -9`
    - Files are stored in an efficient manner, with a shallow directory structure.
- You can restore the situation of the roots at a given snapshot, later on
    - Files' and dirs' mode and modification time are preserved
- If you choose to delete any snapshot, dangling backup files are removed.
- Of course, it's possible to list roots and snapshots and delete any of them.

All paths are converted to absolute paths, for consistency.

## Mini-tutorial

We will backup the contents of the `C:\MyImportantDir`, using the `C:\MySnapkupDir` folder as the container of the backup structures. This example is styled after windows, but it's completely similar under UNIXes.

### Initialize the backup directory

`snapkup.exe -d C:\MySnapkupDir init`

### Register the directory to backup as a root

`snapkup.exe -d C:\MySnapkupDir add-root C:\MyImportantDir`

### Take your first snapshot

`snapkup.exe -d C:\MySnapkupDir snap`

*add `-z` if you want to compress the files being backed up*.

### Delete it, because... just because.

`snapkup.exe -d C:\MySnapkupDir del-snap 0`

### Or restore it!

`snapkup.exe -d C:\MySnapkupDir restore 0 C:\MyRestoreDir`

*the destination directory should be empty.*

## Status

Everything described above should work. **It's still at an early stage of development, so don't trust it with any critical data, yet**. 

Next steps:

- Proper testing framework, for reliability
- Improved documentation
- Mounting a snapshot as a FUSE filesystem
- Proper cross-compiling
- Better error handling
- Better recovery of the data structures from errors

## Build

`cd` to the `src/` dir and `go build`. On UNIX systems you can also use `make build` from the root.

It uses `CGO`, so cross-compilation comes with the usual caveats, and a proper build stack should be installed.