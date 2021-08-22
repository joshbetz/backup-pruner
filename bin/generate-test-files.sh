#!/bin/bash

if [ -z "$1" ]; then
	echo "Usage: $0 <directory>"
	exit 1;
fi

DIR=$(dirname $1)

for hour in {25920..1}; do
	DATE=$(date -v"-${hour}H" "+%y%m%d%H00")
	FILENAME=$(date -v"-${hour}H" "+%Y-%m-%d-%H00")
	echo "$DIR/$1/$FILENAME"
	touch -t "$DATE" "$DIR/$1/$FILENAME"
	sleep 0.1
done
