#!/bin/bash
set -e

# the first argument is the supplied mode (dev or integration)
mode="$1"

if [ -z "$mode" ]; then
  echo "mode is required (dev|integration|prod)"
  exit 1
fi


if [ ! -e "./litestream" ]; then
  # get litestream binary
  curl -L https://github.com/benbjohnson/litestream/releases/download/${LITESTREAM_VERSION} -o litestream.tar.gz
  tar xf litestream.tar.gz && rm litestream.tar.gz
fi

if [ -e "./litestream.tar.gz" ]; then
  rm ./litestream.tar.gz
fi

if [ -d "./${mode}" ]; then
  rm -rf ./${mode}
  mkdir ./${mode}
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

echo "will restore from s3 ..."
echo $CORE_REPLICA_URL
echo $BOOKMARKS_REPLICA_URL
echo $MYDMS_REPLICA_URL

./litestream restore -o ./${mode}/core.db $CORE_REPLICA_URL
./litestream restore -o ./${mode}/bookmarks.db $BOOKMARKS_REPLICA_URL
./litestream restore -o ./${mode}/mydms.db $MYDMS_REPLICA_URL

