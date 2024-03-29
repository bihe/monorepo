#!/bin/bash
set -e

# Restore the database if it does not already exist.
if [ -f "$CONNECTIONSTRING" ]; then
	echo "Database already exists, skipping restore -> $CONNECTIONSTRING"
else
	echo "No database found, restoring from replica if exists -> $CONNECTIONSTRING"
	/opt/litestream/litestream restore -v -if-replica-exists -o "$CONNECTIONSTRING" "${REPLICA_URL}"
fi

# Run litestream with your app as the subprocess.
exec /opt/litestream/litestream replicate -exec "$BASEPATH/$BINARYNAME --basepath=$BASEPATH --port=$PORT --hostname=$HOST"
