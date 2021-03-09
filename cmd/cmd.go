package main

import (
	"log"

	"github.com/schwarzlichtbezirk/exwf"
)

func main() {
	log.Println("starts")
	exwf.Init()
	exwf.Run()
	log.Printf("run")
	go func() {
		exwf.WaitBreak()
		log.Println("shutting down by break begin")
	}()
	exwf.WaitExit()
	exwf.Shutdown()
	log.Println("shutting down complete.")
}

// The End.
