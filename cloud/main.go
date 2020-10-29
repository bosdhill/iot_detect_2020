package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime/pprof"
	"sync"
	//"github.com/bosdhill/iot_detect_2020/sdl"
)

var (
	cpuprofile    = flag.String("cpuprofile", "", "write cpu profile to file")
	serverAddr    = flag.String("server_addr", "192.168.1.121:10000", "The edge server address in the format of host:port")
	appServerAddr = flag.String("app_server_addr", "localhost:4200", "The app server address in the format of host:port")
)

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		go func() {
			log.Println(http.ListenAndServe("localhost:6060", nil))
		}()
		defer pprof.StopCPUProfile()
	}
	log.Println("in main")
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	eCtx := &EdgeContext{ctx: ctx, cancel: cancel}
	ds, err := NewDataStore(eCtx)
	if err != nil {
		panic(err)
	}

	aod, err := NewActionOnDetect(eCtx, *appServerAddr)
	if err != nil {
		panic(err)
	}
	_, err = aod.RegisterEvents(classNamesMap)
	if err != nil {
		panic(err)
	}

	od, err := NewObjectDetection(eCtx, aod)
	if err != nil {
		panic(err)
	}

	cComm, err := NewClientCommunication(eCtx, *serverAddr, ds, od)
	if err != nil {
		panic(err)
	}

	wg.Add(2)
	go func() {
		err := cComm.ServeClient()
		if err != nil {
			panic(err)
		}
	}()
	// tests get
	go func() {
		select {
		case <-ctx.Done():
			err := ds.Get()
			if err != nil {
				panic(err)
			}
		}
	}()
	wg.Wait()
	//sdl.Main()
}
