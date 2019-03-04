#!/bin/bash

set -eux

if [ "$#" -ne 1 ]; then
	echo "usage: $0 run-monitor-log-file"
	exit 1
fi
go build .
./run-monitor-chart $1
python3 -m http.server &
PID=$!
trap "kill $PID" SIGINT SIGTERM
xdg-open "http://localhost:8000"
wait