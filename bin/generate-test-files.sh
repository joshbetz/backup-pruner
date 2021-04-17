#!/bin/bash

if [ -z "$1" ]; then
	echo "Usage: $0 <directory>"
	exit 1;
fi

DIR=$(dirname $1)

for hour in {1..25920}; do
	DATE=$(date -v"-${hour}H" "+%y%m%d%H00")
	touch -t "$DATE" "$DIR/$1/$DATE"
done
