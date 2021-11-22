package main

import (
	"log"

	"github.com/schwarzlichtbezirk/exwf"
)

func main() {
	log.Println("starts")
	exwf.Init()
	exwf.Run()
	log.Println("shutting down complete.")
}

// The End.
