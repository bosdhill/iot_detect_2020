package main

import (
	"flag"
	"log"
	"os"
	"sync"
	"time"
)

var (
	appQueryServerAddr = flag.String("app-query-addr", "localhost:4204", "The app query server address in the format of host:port")
	eodServerAddr     = flag.String("eod-server-addr", "localhost:4201", "The app server address in the format of host:port")
	withRealTimeQuery = flag.Bool("realtime", false, "test real time event filtering")
	withEventQuery    = flag.Bool("query", false, "test event query")
	eventQueryPeriod  = flag.Duration("query-rate", 100 * time.Millisecond, "event query period in ms")
	eventQuerySeconds = flag.Int64("seconds", 6000, "filter events from last number of seconds in event filter")
	testTimeout       = flag.Duration("timeout", 30 * time.Second, "test timeout in seconds")
	logEvents         = flag.Bool("print", false, "print events received")
	logsPath		  = flag.String("logs-path", "logs/", "path to logs directory")
)

func main() {
	flag.Parse()

	file, err := os.OpenFile(*logsPath + "logs.txt", os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)

	var wg = &sync.WaitGroup{}

	if *withRealTimeQuery {
		wg.Add(1)
		go TimedTestEventOnDetect(wg)
	}

	if *withEventQuery {
		wg.Add(1)
		go TimedTestQuery(wg)
	}

	wg.Wait()
}
