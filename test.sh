#!/usr/bin/env bash
if [[ $1 == "query" ]]; then
    echo "$1"
    echo -e "use detections\ndb.dropDatabase()" | mongo
    pushd edge
    go build
    ./edge --with-cloud=false &
    EDGE_PID=$!
    pushd ../application
    go build
    pushd ../client/
    go build
    ./client &
    CLIENT_PID=$!
    popd
    ./application --query=true --timeout 60s --seconds=600
    kill ${EDGE_PID}
    kill ${CLIENT_PID}
fi

if [[ $1 == "realtime" ]]; then
    echo "$1"
    echo -e "use detections\ndb.dropDatabase()" | mongo
    pushd edge
    go build
    ./edge --with-cloud=false &
    EDGE_PID=$!
    pushd ../application
    go build
    pushd ../client/
    go build
    ./client --cont-stream=true --datapath=data/traffic-mini.mp4 &
    CLIENT_PID=$!
    popd
    ./application --realtime=true --timeout 60s
    kill ${EDGE_PID}
    kill ${CLIENT_PID}
fi