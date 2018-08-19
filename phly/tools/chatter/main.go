package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Basic application that generates a stream of values with a small amount of control:
// How long to run, how frequently to generate a value.

// Command line args:
// "-dur seconds" The number of seconds to run the app. Default is infinity.
// "-freq seconds" The number of seconds between chatter. Default is 1.

func main() {
	// Let user ctrl-c to quit the app.
	stop := make(chan struct{})
	go func() {
		finished := make(chan os.Signal, 1)
		signal.Notify(finished, os.Interrupt, syscall.SIGTERM)
		<-finished
		close(stop)
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go run(newCfg(os.Args...), stop, &wg)
	wg.Wait()
}

func run(cfg cfg_t, stop chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	duration := time.NewTimer(cfg.duration)
	frequency := time.NewTicker(cfg.frequency)
	defer duration.Stop()
	defer frequency.Stop()

	counter := 0
	fmt.Println(cfg.id, "-", counter)

	for {
		select {
		case <-stop:
			return
		case <-duration.C:
			return
		case <-frequency.C:
			counter++
			fmt.Println(cfg.id, "-", counter)
		}
	}
}
