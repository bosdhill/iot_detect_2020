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
	serverAddr    = flag.String("server-addr", "192.168.1.121:10000", "The edge server address in the format of host:port")
	appServerAddr = flag.String("app-server-addr", "localhost:4200", "The app server address in the format of host:port")
	withCuda      = flag.Bool("with-cuda", false, "Determines whether cuda is enabled or not")
	matprofile    = flag.Bool("matprofile", false, "displays profile count of gocv.Mat")
)

var classNamesMap = map[string]bool{
	"person":         true,
	"bicycle":        true,
	"car":            true,
	"motorcycle":     true,
	"airplane":       true,
	"bus":            true,
	"train":          true,
	"truck":          true,
	"boat":           true,
	"traffic light":  true,
	"fire hydrant":   true,
	"stop sign":      true,
	"parking meter":  true,
	"bench":          true,
	"bird":           true,
	"cat":            true,
	"dog":            true,
	"horse":          true,
	"sheep":          true,
	"cow":            true,
	"elephant":       true,
	"bear":           true,
	"zebra":          true,
	"giraffe":        true,
	"backpack":       true,
	"umbrella":       true,
	"handbag":        true,
	"tie":            true,
	"suitcase":       true,
	"frisbee":        true,
	"skis":           true,
	"snowboard":      true,
	"sports ball":    true,
	"kite":           true,
	"baseball bat":   true,
	"baseball glove": true,
	"skateboard":     true,
	"surfboard":      true,
	"tennis racket":  true,
	"bottle":         true,
	"wine glass":     true,
	"cup":            true,
	"fork":           true,
	"knife":          true,
	"spoon":          true,
	"bowl":           true,
	"banana":         true,
	"apple":          true,
	"sandwich":       true,
	"orange":         true,
	"broccoli":       true,
	"carrot":         true,
	"hot dog":        true,
	"pizza":          true,
	"donut":          true,
	"cake":           true,
	"chair":          true,
	"couch":          true,
	"potted plant":   true,
	"bed":            true,
	"dining table":   true,
	"toilet":         true,
	"tv":             true,
	"laptop":         true,
	"mouse":          true,
	"remote":         true,
	"keyboard":       true,
	"cell phone":     true,
	"microwave":      true,
	"oven":           true,
	"toaster":        true,
	"sink":           true,
	"refrigerator":   true,
	"book":           true,
	"clock":          true,
	"vase":           true,
	"scissors":       true,
	"teddy bear":     true,
	"hair drier":     true,
	"toothbrush":     true}

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

	var wg sync.WaitGroup
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
