package main

import (
	"log"
	"sync"
	"context"

	//"github.com/bosdhill/iot_detect_2020/sdl"
)

func main() {
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
