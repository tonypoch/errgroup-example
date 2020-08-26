package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx, done := context.WithTimeout(context.Background(), 10*time.Second)
	eg, childctx := errgroup.WithContext(ctx)

	c := make(chan bool)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	// closing
	eg.Go(func() error {
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
		select {
		case sig := <-signalChannel:
			log.Printf("Received signal: %s\n", sig)
			done()
		case <-childctx.Done():
			log.Println("Closing signal goroutine")
			return childctx.Err()
		}
		return nil
	})

	eg.Go(func() error {
		for {
			select {
			case <-childctx.Done():
				log.Println("Closing func1")
				return nil
			case received := <-c:
				if received {
					time.Sleep(100 * time.Millisecond)
					log.Println("func1 is over!")
				}
			}
		}
	})
	//func 2
	var fl bool
	eg.Go(func() error {
		t0 := time.Now()
		for {
			select {
			case <-childctx.Done():
				log.Println("Closing func2")
				return nil
			case t := <-ticker.C:
				log.Println("tick!")
				if t.Sub(t0) > 5*time.Second {
					err := fmt.Errorf("time out!")
					log.Println(err)
					return err
				}
				fl = !fl
				c <- fl
			}
		}
	})

	if err := eg.Wait(); err != nil {
		if errors.Is(err, context.Canceled) {
			log.Println("Context was canceled")
		} else {
			log.Printf("Received error: %v\n", err)
		}
	} else {
		log.Println("Finished clean")
	}
}
