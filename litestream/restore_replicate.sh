#!/bin/bash
set –e

database_connections=(
    "/opt/litestream/store/core.db"
    "/opt/litestream/store/bookmarks.db"
    "/opt/litestream/store/mydms.db"
)

for con in "${database_connections[@]}"; do
	# Restore the database if it does not already exist.
	if [ -f "$con" ]; then
		echo "Database already exists, skipping restore -> $con"
	else
		echo "No database found, restoring from replica if exists -> $con"
		/opt/litestream/litestream restore -if-replica-exists "$con"
	fi
done

# start the litestream replication
/opt/litestream/litestream replicate
