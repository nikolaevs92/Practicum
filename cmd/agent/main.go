package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nikolaevs92/Practicum/internal/agent"
	"github.com/nikolaevs92/Practicum/internal/config"
)

func main() {
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-cancelChan
		cancel()
	}()

	conf := config.LoadConfig()
	collector := agent.New(*conf.Agent)
	collector.Run(ctx)

	log.Println("Program end")
}
