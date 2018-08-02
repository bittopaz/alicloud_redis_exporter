#!/usr/bin/env bash

if [[ -n ${APP_PATH} ]]; then
    OLD_PATH=$PWD
    cd $APP_PATH
fi
pwd
# Check gofmt
echo "==> Checking that code complies with gofmt requirements..."
gofmt_files=$(gofmt -l `find . -name '*.go' | grep -v vendor`)
if [[ -n ${gofmt_files} ]]; then
    echo 'gofmt needs running on the following files:'
    echo "${gofmt_files}"
    echo "You can use the command: \`make fmt\` to reformat code."
    exit 1
fi

if [[ -n ${OLD_PATH} ]]; then
    cd $OLD_PATH
fi

exit 0
