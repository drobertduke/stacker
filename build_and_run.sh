#!/bin/sh
pkill stacker
go build
if [ $? -eq 0 ]
then
	echo "BUILD SUCCESS"
fi
