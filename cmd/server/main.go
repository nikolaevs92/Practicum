package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nikolaevs92/Practicum/internal/config"
	"github.com/nikolaevs92/Practicum/internal/server"
)

func main() {
	cfg := config.LoadConfig()
	dataServer := server.New(*cfg.Server)
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-cancelChan
		cancel()
	}()
	dataServer.Run(ctx)

	fmt.Println("Program end")
}
