package main

import (
	"flag"
	"sync"
)

var (
	appQueryServerAddr = flag.String("app-query-addr", "localhost:4204", "The app query server address in the format of host:port")
	eodServerAddr      = flag.String("eod-server-addr", "localhost:4201", "The app server address in the format of host:port")
)

func main() {
	flag.Parse()

	var wg sync.WaitGroup
	wg.Add(2)

	go TestEventOnDetect(&wg)
	go TestQuery(&wg)

	wg.Wait()
}
