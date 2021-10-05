- **init** \<dir\>
	- verify that dir is empty
	- create the database
	- create the "baseline" snap database
- **add-root** \<root\>
	- open the database
	- makes the specified root absolute
	- adds a root if it doesn't exist
- **list-roots**
	- open the database
	- gathers and prints the roots
- **snap**
	- open the database and a transaction
	- determine the snap ID and create a new snap
	- list of files in roots
	- list of files in the latest snap
	- compare the two lists to get:
		- AU: added or updated files
		- NC: files not changed
	- for NC, set max_gen to new snap ID
	- for AU, create new record with min_gen, max_gen and stored_gen set to new snap ID
	- create a new snap database and import in it the AU files
	- commit
- **list-snaps**
	- open the database
	- gathers and prints the snaps, with date, ordered by date
- **prune** \<snap\>
	- open the database and a transaction
	- check that the snap is not >= current snap
	- select the files to delete (where max_gen <= snap)
		- delete them from the snap databases
	- select the files to copy in the baseline (where storage_gen <= snap)
		- copy them to baseline database
		- mark storage_gen as -1
		- mark min_gen as snap + 1 where it is <= snap 
- **restore** \<snap\> \<destination_empty_dir\>
	- open the database, RO
	- select files where min_gen <= snap <= max_gen
	- open the snap databases and go thru them restoring the files
```

```
   ____               __           
  / __/__  ___  ___  / /_ __ _ ___ 
 _\ \/ _ \/ _ `/ _ \/  '_/ // / _ \
/___/_//_/\_,_/ .__/_/\_\\_,_/ .__/
             /_/   v0.0.1   /_/
```