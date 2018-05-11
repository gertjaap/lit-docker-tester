#!/bin/bash
set -e

docker build ~/go/src/github.com/mit-dci/lit -t lit

./run.sh $1
