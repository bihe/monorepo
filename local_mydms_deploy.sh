#!/bin/sh

TSTAMP=$(date -u +%Y%m%d.%H%M%S)
COMMIT=`git rev-parse HEAD | cut -c 1-8`

make clean build

rm -rf ./assets/bundle
cd ./assets
PATH=./node_modules/.bin:$PATH ./bundle.sh auto ${COMMIT}
cd ..
cp ./dist/linux/arm64/* ../mydms.deployment
cp -R ./assets/ ../mydms.deployment
rm ../mydms.deployment/assets/.gitignore
cd ../mydms.deployment
git add .
git commit -m "new deployment version ${TSTAMP}"
git push
