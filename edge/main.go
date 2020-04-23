package main

import (
	"log"
	"sync"
	sdl "github.com/bosdhill/iot_detect_2020/sdl"
)

func main() {
	log.Println("in main")
	var wg sync.WaitGroup
	wg.Add(1)
	ServeClient()
	wg.Wait()
	sdl.Beep()
}
