package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"go.uber.org/automaxprocs/maxprocs"
)

var build = "develop"

func main() {

	// =========================================================================
	// GOMAXPROCS

	// Set the correct number of threads for the service
	// based on what is available either by the machine or quotas.
	if _, err := maxprocs.Set(); err != nil {
		fmt.Println("maxprocs: ", err)
		os.Exit(1)
	}
	g := runtime.GOMAXPROCS(0)

	log.Printf("staring service build: [%s]; CPUs: [%d]", build, g)
	defer log.Println("service ended")

	shutDown := make(chan os.Signal, 1)
	signal.Notify(shutDown, syscall.SIGINT, syscall.SIGTERM)
	<-shutDown

	log.Println("stopping service")
}
