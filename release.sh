#! /bin/bash

UNCOMMITTED="yes"

VERSION=0.2.2

if [ "x$VERSION" == "x" ]; then
  echo "Please export VERSION=X.X.X"
  exit 1
fi

if [ "x$GITHUB_TOKEN" == "x" ]; then
  echo "Please export GITHUB_TOKEN=...."
  exit 1
fi

MYDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
BASE=$MYDIR

TEST=$(git status --porcelain|wc -l)
if [ 0 -ne $TEST -a $UNCOMMITTED != "yes" ]; then
   echo "Please, commit before releasing"
   exit 1
fi

echo "Let's go"
git tag $VERSION

goreleaser release --clean --skip-validate



