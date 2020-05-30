package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"runtime/pprof"
	"sync"
	"context"
	_ "net/http/pprof"
	//"github.com/bosdhill/iot_detect_2020/sdl"
)
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

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

	od, err := NewObjectDetection(eCtx)
	if err != nil {
		panic(err)
	}

	cComm, err := NewClientCommunication(eCtx, ds, od)
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
