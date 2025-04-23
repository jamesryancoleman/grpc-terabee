package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jamesryancoleman/terabee"
)

const defaultAddr string = "0.0.0.0:50065"

func main() {
	terabee.StartServer(defaultAddr)

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("server running, press ctrl+c to exit...")
	<-done // Will block here until user hits ctrl+c
	fmt.Println("\nshutting down.")
}
