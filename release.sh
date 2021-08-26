#!/bin/bash

export GOOS='linux'
ALL_ARC=( "amd64" "386" "arm" "arm64" )

NAME='uWSGI_exporter'
VERSION=`git describe --abbrev=0 --tags`
RELEASE_FOLDER="release"

# Clean OLD release folder if preset and recreate it
rm -rfv $RELEASE_FOLDER
mkdir -p $RELEASE_FOLDER

for arc in "${ALL_ARC[@]}"
do
    export GOARCH=$arc

    FINAL_FOLDER_NAME="$NAME-$VERSION.$GOOS-$GOARCH"
    ARCHIVE_NAME="$FINAL_FOLDER_NAME.tar.gz"

    make clean

    make
    mkdir -p $FINAL_FOLDER_NAME
    mv -v $NAME $FINAL_FOLDER_NAME
    tar -zcvf $ARCHIVE_NAME $FINAL_FOLDER_NAME
    mv $ARCHIVE_NAME $RELEASE_FOLDER
    rm -rfv $FINAL_FOLDER_NAME
done
