package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nikolaevs92/Practicum/internal/config"
	"github.com/nikolaevs92/Practicum/internal/server"
)

func main() {
	os.Setenv("STORE_FILE", "./.data")
	cfg := config.LoadConfig()
	dataServer := server.New(*cfg.Server)
	cancelChan := make(chan os.Signal, 1)
	// tick := time.NewTicker(4 * time.Second)

	signal.Notify(cancelChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-cancelChan
		// <-tick.C
		cancel()
	}()
	dataServer.Run(ctx)

	log.Println("Program end")
}
