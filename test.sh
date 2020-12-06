#!/usr/bin/env bash
NUM_TEST_RUNS=10
NUM_UPLOAD_TEST_RUNS=5

function print_db_stats {
    MB=1024*1024
    GB=1024*${MB}
    echo "MONGODB STATS"
    printf "size (GB): "
    echo -e "use detections\ndb.detection_result.stats()[\"size\"]/($GB)" | mongo --quiet | sed '1d'
    printf "count: "
    echo -e "use detections\ndb.detection_result.stats()[\"count\"]" | mongo --quiet | sed '1d'
    printf "avgObjSize (MB):"
    echo -e "use detections\ndb.detection_result.stats()[\"avgObjSize\"]/($MB)" | mongo --quiet | sed '1d'
}

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
    print_db_stats
}

function test_query_realtime_db_delete_upload {
    echo "TEST_QUERY_REALTIME_DB_DELETE_UPLOAD"
    EDGE_FLAGS="-with-cloud=true --batchsize=100 --uploadTTL=30 --deleteTTL=60"
    CLIENT_FLAGS="--cont-stream=true"
    APP_FLAGS="--realtime=true --query=true --timeout 120s --seconds=60"
    echo "flags: ${EDGE_FLAGS} ${CLIENT_FLAGS} ${APP_FLAGS}"
    echo -e "use detections\ndb.dropDatabase()" | mongo &> /dev/null
    for (( c=1; c<=$NUM_UPLOAD_TEST_RUNS; c++ ))
    do
        test_base "${EDGE_FLAGS}" "${CLIENT_FLAGS}" "${APP_FLAGS}"
    done
}

function test_query_db_delete_upload {
    echo "TEST_QUERY_DB_DELETE_UPLOAD"
    EDGE_FLAGS="-with-cloud=true --batchsize=100 --uploadTTL=30 --deleteTTL=60"
    CLIENT_FLAGS="--cont-stream=true"
    APP_FLAGS="--query=true --timeout 120s --seconds=60"
    echo "flags: ${EDGE_FLAGS} ${CLIENT_FLAGS} ${APP_FLAGS}"
    echo -e "use detections\ndb.dropDatabase()" | mongo &> /dev/null
    for (( c=1; c<=$NUM_UPLOAD_TEST_RUNS; c++ ))
    do
        test_base "${EDGE_FLAGS}" "${CLIENT_FLAGS}" "${APP_FLAGS}"
    done
}

function test_realtime_db_delete_upload {
    echo "TEST_REALTIME_DB_DELETE_UPLOAD"
    EDGE_FLAGS="-with-cloud=true --batchsize=100 --uploadTTL=30 --deleteTTL=60"
    CLIENT_FLAGS="--cont-stream=true"
    APP_FLAGS="--realtime=true --timeout 120s"
    echo "flags: ${EDGE_FLAGS} ${CLIENT_FLAGS} ${APP_FLAGS}"
    echo -e "use detections\ndb.dropDatabase()" | mongo &> /dev/null
    for (( c=1; c<=$NUM_UPLOAD_TEST_RUNS; c++ ))
    do
        test_base "${EDGE_FLAGS}" "${CLIENT_FLAGS}" "${APP_FLAGS}"
    done
}

function test_query_realtime_db_delete {
    echo "TEST_QUERY_REALTIME_DB_DELETE"
    EDGE_FLAGS="-with-cloud=false"
    CLIENT_FLAGS="--cont-stream=true"
    APP_FLAGS="--realtime=true --query=true --timeout 60s --seconds=60"
    echo "flags: ${EDGE_FLAGS} ${CLIENT_FLAGS} ${APP_FLAGS}"
    echo -e "use detections\ndb.dropDatabase()" | mongo &> /dev/null
    for (( c=1; c<=$NUM_TEST_RUNS; c++ ))
    do
        test_base "${EDGE_FLAGS}" "${CLIENT_FLAGS}" "${APP_FLAGS}"
    done
}

function test_query_db_delete {
    echo "TEST_QUERY_DB_DELETE"
    EDGE_FLAGS="-with-cloud=false"
    CLIENT_FLAGS="--cont-stream=true"
    APP_FLAGS="--query=true --timeout 60s --seconds=60"
    echo "flags: ${EDGE_FLAGS} ${CLIENT_FLAGS} ${APP_FLAGS}"
    echo -e "use detections\ndb.dropDatabase()" | mongo &> /dev/null
    for (( c=1; c<=$NUM_TEST_RUNS; c++ ))
    do
        test_base "${EDGE_FLAGS}" "${CLIENT_FLAGS}" "${APP_FLAGS}"
    done
}

function test_realtime_db_delete {
    echo "TEST_REALTIME_DB_DELETE"
    EDGE_FLAGS="-with-cloud=false"
    CLIENT_FLAGS="--cont-stream=true"
    APP_FLAGS="--realtime=true --timeout 60s"
    echo "flags: ${EDGE_FLAGS} ${CLIENT_FLAGS} ${APP_FLAGS}"
    echo -e "use detections\ndb.dropDatabase()" | mongo &> /dev/null
    for (( c=1; c<=$NUM_TEST_RUNS; c++ ))
    do
        test_base "${EDGE_FLAGS}" "${CLIENT_FLAGS}" "${APP_FLAGS}"
    done
}

function test_query_metadata_db_delete {
    echo "TEST_QUERY_METADATA_DB_DELETE"
    EDGE_FLAGS="-with-cloud=false"
    CLIENT_FLAGS="--cont-stream=true"
    APP_FLAGS="--query=true --timeout 60s --seconds=60 --metadata=true"
    echo "flags: ${EDGE_FLAGS} ${CLIENT_FLAGS} ${APP_FLAGS}"
    echo -e "use detections\ndb.dropDatabase()" | mongo &> /dev/null
    for (( c=1; c<=$NUM_TEST_RUNS; c++ ))
    do
        test_base "${EDGE_FLAGS}" "${CLIENT_FLAGS}" "${APP_FLAGS}"
    done
}

function test_query_db_persist {
    echo "TEST_QUERY_DB_PERSIST"
    EDGE_FLAGS="-with-cloud=false"
    CLIENT_FLAGS="--cont-stream=true"
    APP_FLAGS="--query=true --timeout 60s --seconds=60"
    echo "flags: ${EDGE_FLAGS} ${CLIENT_FLAGS} ${APP_FLAGS}"
    for (( c=1; c<=$NUM_TEST_RUNS; c++ ))
    do
        test_base "${EDGE_FLAGS}" "${CLIENT_FLAGS}" "${APP_FLAGS}"
    done
}

function test_realtime_db_persist {
    echo "TEST_REALTIME_DB_PERSIST"
    EDGE_FLAGS="-with-cloud=false"
    CLIENT_FLAGS="--cont-stream=true"
    APP_FLAGS="--query=true --timeout 60s"
    echo "flags: ${EDGE_FLAGS} ${CLIENT_FLAGS} ${APP_FLAGS}"
    for (( c=1; c<=$NUM_TEST_RUNS; c++ ))
    do
        test_base "${EDGE_FLAGS}" "${CLIENT_FLAGS}" "${APP_FLAGS}"
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
    "realtime_db_delete_upload")
        test_realtime_db_delete_upload
        ;;
    "query_db_delete_upload")
        test_query_db_delete_upload
        ;;
    "query_realtime_db_delete_upload")
        test_query_realtime_db_delete_upload
        ;;
esac
