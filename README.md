# 🗃️ snapkup v0.3.0

Snapkup is a simple backup tool that takes snapshots of your filesystem (or the parts that you'll decide), storing them
efficiently and conveniently.

## Basic workflow

The basic flow is:

- You initialize an empty directory that will store the backups
- You register one or more backup roots, directory or files that will be snapshotted
- You take one or more snapshots. Snapkup lists all the tree for those roots, taking a copy of the contents
- You can restore the situation of the roots at any given snapshot
- Of course, it's possible to list roots and snapshots and delete any of them, and perform all the other admin ops 

Notable points:

- Files are deduplicated: only one copy of a file is stored, across the filesystem and all the snapshots
- Everything stored on-disk is encrypted, using `XChaCha20Poly1305`
- Checksums, using authenticated 128-bit `SipHash`, are used to perform deduplication and integrity
- By default, everything is compressed using `zstd -19`. Incompressible files are stored as not compressed.
- Small files can be merged in "agglos", to reduce the number of files and make it more sync-friendly (e.g. for Dropbox)
- Snapkup favors features and code readability over speed. It's not slow, though!
- All paths are converted to absolute paths, for consistency.
- Cross-platform portability of backup archives is not a priority, though it should reasonably work.

Plans for the future:

- Ability to produce all outputs as JSON, for better script-ability
- Ability to retrieve files from external filesystems, via SSH
- Ability to back up data that come from the execution of a command (e.g. `crontab -l`)
- FUSE-mount a snapshot

## Mini-tutorial

We will back up the contents of the `C:\MyImportantDir`, using the `C:\MySnapkupDir` folder as the container of the 
backup structures. This example is styled after windows, but it's completely similar under UNIXes.

**N.B.**: all the commands have shortcuts; e.g. `root add` can be `r a`. Read the help (`snapkup --help`)

**N.B.**: UNIX-style command is used (`snapkup`); of course, under windows you can use `snapkup.exe`

### Set the encryption password

For now, it's read from an environment variable, `SNAPKUP_PASSWORD`, so you can use:

```
[UNIX/BASH]      export SNAPKUP_PASSWORD=MyCoolPwd
[WIN/CMD]        set "SNAPKUP_PASSWORD=MyCoolPwd"
[WIN/POWERSHELL] $env:SNAPKUP_PASSWORD = 'MyCoolPwd'
```

### Initialize the backup directory

You need to initialize a directory to store the backups to. It's specified with the `--backup-dir` or `-d` flags, and 
this flag will need to be repeated for every command.

`snapkup -d C:\MySnapkupDir init`

It requires an empty directory and creates a shallow dir structure to organize the files, and a `snapkup.dat` file (encrypted) 
that will store all the metadata. Also, generates the encryption and checksum keys.

### Register the directory to back up as a root

`snapkup -d C:\MySnapkupDir root add C:\MyImportantDir`

Adds a directory as one of the paths to back up. All its contents will be recursively scanned when performing a snapshot 
(see below).

It can also be a single file. The absolute path will be stored, to avoid ambiguities.

As many roots as you want can be stored; `root list` and `root del` are available to manage the list.

### Take your first snapshot

`snapkup -d C:\MySnapkupDir snap do`

It walks the roots' filesystem trees, and hashes every file. It then compares the hashes with the files already stored, 
and stores only those files that are not already seen.

`snapkup -d C:\MySnapkupDir snap do -l "My first label"`

All (unique) files are stored as data "blobs", that are compressed (unless `--no-compress` is specified), encrypted and 
protected with a checksum.

Metadata (path, mod time, access mode) of files and dirs is preserved for each snapshot.

A snap ID is returned, that can be used for a variety of operations: `snap label`, `snap list`, `snap filelist`, 
`snap info` and of course `snap del`.

Removing a snapshot with `snap del` removes all the orphaned blobs, freeing disk space.

### Merge small files

When having a multitude of small files is not desirable, e.g. in a remote sync scenario, it's possible to merge files 
in an "agglo". You can specify the threshold size of the files to merge and the target size of the agglo, in megabytes.

`agglo calc` allows you to evaluate the number of files that will be merged, and the result.

`snapkup -d C:\MySnapkupDir agglo calc 1 5`

This will merge all the files up to 1Mb in agglos that are (about) 5Mb. Use `agglo do` with the same parameters, to
actually perform the merge.

`snapkup -d C:\MySnapkupDir agglo unpack`

Does the opposite, unmerging the files and removing the agglos.

**N.B.** when deleting a snapshot that references a blob inside an agglo, the agglo is not modified even if it's the
last reference, to avoid triggering a sync. To reclaim the space, `unpack` the agglos and the dangling files will not
be restored; then `perform` again.

### Restore it!

To restore all the roots for snapshot `0`:

`snapkup -d C:\MySnapkupDir snap restore 0 C:\MyRestoreDir`

The destination directory must be empty.

It is also possible to specify a prefix path to select only a part of the file list:

`snapkup -d C:\MySnapkupDir snap restore 0 C:\MyRestoreDir --prefix-path /foo/bar`

## Status

Everything described above should work. **It's still at an early stage of development, so don't trust it with any 
critical data, yet**. 

Next steps:

- Further unit testing
- Improve documentation
  - Document the on-disk layout of files, for external review
- Better error handling
- Better recovery of the data structures from errors
- Better/more convenient handling of passwords

## Build

`cd` to the `src/` dir and `go build`. On UNIX systems you can also use `make build` from the root.

It uses `CGO`, so cross-compilation comes with the usual caveats, and a proper build stack should be installed.