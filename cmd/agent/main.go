package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nikolaevs92/Practicum/internal/agent"
	"github.com/nikolaevs92/Practicum/internal/config"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-cancelChan
		cancel()
	}()

	adress := pflag.String("a", config.DefaultServer, "")
	pollInterval := pflag.Duration("p", config.DefaultPollInterval, "")
	reportInterval := pflag.Duration("r", config.DefaultReportInterval, "")
	pflag.Parse()

	v := viper.New()
	conf := config.NewAgentConfigWithDefaults(v, *adress, *pollInterval, *reportInterval)
	collector := agent.New(*conf)
	collector.Run(ctx)

	log.Println("Program end")
}
