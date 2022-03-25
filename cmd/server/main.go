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
	log.Println("FINDME")
	adress := pflag.StringP("adress", "a", config.DefaultServer, "")
	storeInterval := pflag.DurationP("strore-interval", "i", config.DefaultStoreInterval, "")
	storeFile := pflag.StringP("store-file", "f", config.DefaultStoreFile, "")
	restore := pflag.BoolP("restore", "r", config.DefaultRestore, "")
	key := pflag.StringP("key", "k", "", "")
	dataBaseDSN := pflag.StringP("db-dsn", "d", "", "")
	dataBaseType := pflag.StringP("db-type", "t", "", "")
	pflag.Parse()

	v := viper.New()
	v.AllowEmptyEnv(true)
	v.AutomaticEnv()

	cfg := config.NewServerConfigWithDefaults(v, *adress, *storeInterval, *storeFile, *restore, *key, *dataBaseDSN, *dataBaseType)
	log.Println("DSN: " + cfg.DataBaseDSN)
	log.Println("server: " + cfg.Server)
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
