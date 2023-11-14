#!/bin/sh

if [ ! -d ".core.db-litestream" ]; then
  rm -rf ./.core.db-litestream
fi

if [ ! -d ".mydms.db-litestream" ]; then
  rm -rf ./.mydms.db-litestream
fi

if [ ! -d ".bookmarks.db-litestream" ]; then
  rm -rf ./.bookmarks.db-litestream
fi

rm mydms*
rm core*
rm bookmarks*

litestream restore -o ./core.db $CORE_REPLICA_URL
litestream restore -o ./bookmarks.db $BOOKMARKS_REPLICA_URL
litestream restore -o ./mydms.db $MYDMS_REPLICA_URL
