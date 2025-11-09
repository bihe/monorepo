#!/bin/bash
set –e

database_connections=(
    "/opt/litestream/store/core.db"
    "/opt/litestream/store/bookmarks.db"
    "/opt/litestream/store/mydms.db"
)

for con in "${database_connections[@]}"; do
	if [ -f "$con" ]; then
		echo "Database already exists -> $con"
	else
		echo "No database found -> $con"
		exit 1
	fi
done

exit 0
