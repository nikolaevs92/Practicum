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

	address := pflag.StringP("address", "a", config.DefaultServer, "")
	pollInterval := pflag.DurationP("pool-inreval", "p", config.DefaultPollInterval, "")
	reportInterval := pflag.DurationP("report-interval", "r", config.DefaultReportInterval, "")
	key := pflag.StringP("key", "k", "", "")
	pflag.Parse()

	v := viper.New()
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	conf := config.NewAgentConfigWithDefaults(v, *address, *pollInterval, *reportInterval, *key)
	collector := agent.New(*conf)
	collector.Run(ctx)

	log.Println("Program end")
}
