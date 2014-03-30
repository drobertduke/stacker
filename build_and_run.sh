#!/bin/sh
pkill stacker
go build
echo "BUILD SUCCESS"
./stacker &
