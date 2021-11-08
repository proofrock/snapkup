# 🗃️ snapkup v0.3.2

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

## Documentation

Please see the [GitBook Documentation](https://germ.gitbook.io/snapkup/) that sports a tutorial and the documentation for all the commands!

## Status

Everything described in the documentation should work. **It's still at an early stage of development, so don't trust it with any 
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
