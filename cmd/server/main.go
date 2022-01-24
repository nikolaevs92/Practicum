package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nikolaevs92/Practicum/internal/config"
	"github.com/nikolaevs92/Practicum/internal/server"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func main() {
	adress := pflag.StringP("adress", "a", config.DefaultServer, "")
	storeInterval := pflag.DurationP("strore-interval", "i", config.DefaultStoreInterval, "")
	storeFile := pflag.StringP("store-file", "f", config.DefaultStoreFile, "")
	restore := pflag.BoolP("restore", "r", config.DefaultRestore, "")
	pflag.Parse()

	v := viper.New()
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	cfg := config.NewServerConfigWithDefaults(v, *adress, *storeInterval, *storeFile, *restore)
	dataServer := server.New(*cfg)
	cancelChan := make(chan os.Signal, 1)

	signal.Notify(cancelChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-cancelChan
		cancel()
	}()
	dataServer.Run(ctx)

	log.Println("Program end")
}
