#!/bin/sh

if [ $# -ne 1 ]; then
    echo "Usage: $0 <number_of_lines>" >&2
    exit 1
fi

i=1
while [ $i -le "$1" ]; do
    echo "$i SOME TEXT"
    i=$((i + 1))
done
