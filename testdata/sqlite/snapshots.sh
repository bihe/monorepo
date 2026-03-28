#!/bin/bash
set -e

# the first argument is the supplied mode (dev or integration)
mode="$1"

if [ -z "$mode" ]; then
  echo "mode is required (dev|integration)"
  exit 1
fi

# the second argument is the database (core.db, bookmarks.db, mydms.db)
db="$2"

if [ -z "$db" ]; then
  echo "db is required (core.db|bookmarks.db|mydms.db)"
  exit 1
fi

. ./.env

if [[ -z "${LITESTREAM_ACCESS_KEY_ID}" ]]; then
  echo "LITESTREAM_ACCESS_KEY_ID is required"
  exit 1
fi

if [[ -z "${LITESTREAM_SECRET_ACCESS_KEY}" ]]; then
  echo "LITESTREAM_SECRET_ACCESS_KEY is required"
  exit 1
fi

CORE_REPLICA_URL="${CORE_REPLICA_URL//__MODE__/${mode}}"
BOOKMARKS_REPLICA_URL="${BOOKMARKS_REPLICA_URL//__MODE__/${mode}}"
MYDMS_REPLICA_URL="${MYDMS_REPLICA_URL//__MODE__/${mode}}"

if [ -e "./litestream.yml" ]; then
  rm ./litestream.yml
fi

# generate the litestream.yaml file
cat > ./litestream.yml <<END
dbs:
  - path: ./${mode}/core.db
    replicas:
      - url: ${CORE_REPLICA_URL}

  - path: ./${mode}/bookmarks.db
    replicas:
      - url: ${BOOKMARKS_REPLICA_URL}

  - path: ./${mode}/mydms.db
    replicas:
      - url: ${MYDMS_REPLICA_URL}
END


# get the available snapshots of the database
./litestream snapshots -config litestream.yml ./${mode}/${db}

rm ./litestream.yml
