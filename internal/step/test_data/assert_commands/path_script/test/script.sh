#!/bin/bash

echo -n "$0 $@"
if [[ $0 == ./* ]]; then
    echo "relative"
    exit 0
elif [[ $0 == /* ]]; then
    echo "absolute"
    exit 0
fi
exit 1
