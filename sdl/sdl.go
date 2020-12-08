// Ref: https://github.com/golang/go/wiki/LockOSThread
package sdl

import (
	"gocv.io/x/gocv"
	"log"
	"runtime"
)

// Arrange that main.main runs on main thread.
func init() {
	runtime.LockOSThread()
}

// Main runs the main SDL service loop.
// The binary's main.main must call sdl.Main() to run this loop.
// Main does not return. If the binary needs to do other work, it
// must do it in separate goroutines.
func Main() {
	for f := range mainfunc {
		f()
	}
}

// queue of work to run in main thread.
var mainfunc = make(chan func())

// do runs f on the main thread.
func do(f func()) {
	done := make(chan bool, 1)
	mainfunc <- func() {
		f()
		done <- true
	}
	<-done
}

func Show(window *gocv.Window, mat *gocv.Mat) {
	log.Println("Show")
	do(func() {
		window.IMShow(*mat)
		window.WaitKey(1)
	})
}
