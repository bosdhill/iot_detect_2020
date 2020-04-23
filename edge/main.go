package main

import (
	"log"
	"sync"

	//"github.com/bosdhill/iot_detect_2020/sdl"
)

func main() {
	log.Println("in main")
	var wg sync.WaitGroup
	wg.Add(1)
	go ServeClient()
	wg.Wait()
	//sdl.Main()
}
