package main

import (
	"log"
	"sync"

	//"github.com/bosdhill/iot_detect_2020/sdl"
)

func main() {
	log.Println("in main")
	var wg sync.WaitGroup
	ds, err := NewDataStore()
	if err != nil {
		panic(err)
	}
	cComm, err := NewClientCommunication(ds)
	wg.Add(1)
	go func() {
		err := cComm.ServeClient()
		if err != nil {
			panic(err)
		}
	}()
	wg.Wait()
	//sdl.Main()
}
