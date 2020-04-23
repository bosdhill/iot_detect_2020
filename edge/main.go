package main

import (
	"log"
	"sync"
)

func main() {
	log.Println("in main")
	var wg sync.WaitGroup
	wg.Add(1)
	go ServeClient()
	wg.Wait()
}
