#!/usr/bin/env bash
NUM_TEST_RUNS=10

function test_base {
    pushd edge &> /dev/null
    go build
    ./edge $1 &
    EDGE_PID=$!
    pushd ../application &> /dev/null
    go build
    pushd ../client/ &> /dev/null
    go build
    ./client $2 &
    CLIENT_PID=$!
    popd  &> /dev/null
    ./application $3
    popd &> /dev/null
    tac logs/logs.txt | awk '/Object_Detection_Time_AVG/ {print;exit}'
    rm logs/logs.txt
    disown ${EDGE_PID}
    kill -9 ${EDGE_PID} &> /dev/null
    disown ${CLIENT_PID}
    kill -9 ${CLIENT_PID} &> /dev/null
    popd &> /dev/null
}

function test_query_db_delete {
    echo "TEST_QUERY_DB_DELETE"
    echo -e "use detections\ndb.dropDatabase()" | mongo &> /dev/null
    for (( c=1; c<=$NUM_TEST_RUNS; c++ ))
    do
        test_base "-with-cloud=false" "--cont-stream=true" \
                  "--query=true --timeout 60s --seconds=60"
    done
}

function test_query_metadata_db_delete {
    echo "TEST_QUERY_METADATA_DB_DELETE"
    echo -e "use detections\ndb.dropDatabase()" | mongo &> /dev/null
    for (( c=1; c<=$NUM_TEST_RUNS; c++ ))
    do
        test_base "-with-cloud=false" "--cont-stream=true" \
                  "--query=true --timeout 60s --seconds=600 --metadata=true"
    done
}

function test_realtime_db_delete {
    echo "TEST_REALTIME_DB_DELETE"
    echo -e "use detections\ndb.dropDatabase()" | mongo &> /dev/null
    for (( c=1; c<=$NUM_TEST_RUNS; c++ ))
    do
        test_base "-with-cloud=false" "--cont-stream=true" \
                  "--realtime=true --timeout 60s"
    done
}

function test_query_db_persist {
    echo "TEST_QUERY_DB_PERSIST"
    for (( c=1; c<=$NUM_TEST_RUNS; c++ ))
    do
        test_base "-with-cloud=false" "--cont-stream=true" \
                  "--query=true --timeout 60s --seconds=600"
    done
}

function test_realtime_db_persist {
    echo "TEST_REALTIME_DB_PERSIST"
    for (( c=1; c<=$NUM_TEST_RUNS; c++ ))
    do
        test_base "-with-cloud=false" "--cont-stream=true" \
                  "--query=true --timeout 60s --seconds=600"
    done
}

function test_query_realtime_db_delete {
    echo "TEST_QUERY_REALTIME_DB_DELETE"
    echo -e "use detections\ndb.dropDatabase()" | mongo &> /dev/null
    for (( c=1; c<=$NUM_TEST_RUNS; c++ ))
    do
        test_base "-with-cloud=false" "--cont-stream=true" \
                  "--realtime=true --query=true --timeout 60s --seconds=60"
    done
}

case "$1" in
     "query_db_delete")
        test_query_db_delete
        ;;
     "query_metadata_db_delete")
        test_query_metadata_db_delete
        ;;
     "query_db_persist")
        test_query_db_persist
        ;;
     "realtime_db_persist")
        test_realtime_db_persist
        ;;
     "realtime_db_delete")
        test_realtime_db_delete
        ;;
     "query_realtime_db_delete")
        test_query_realtime_db_delete
        ;;
esac
