package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/viper"

	"github.com/krtffl/torro/internal/api"
	"github.com/krtffl/torro/internal/config"
	"github.com/krtffl/torro/version"
)

var configPath = flag.String(
	"config",
	"config/config.yaml",
	"path from where the config file will be loaded",
)

func main() {
	flag.Parse()
	banner()

	cfg := config.Load(viper.New(), *configPath)
	api := api.New(cfg)

	go api.Run()
	handleSignals(api)
}

func banner() {
	fmt.Printf("--------------------------------\n")
	fmt.Printf("          el torró torró        \n")
	fmt.Printf("--------------------------------\n")
	fmt.Printf(" Version  : %s\n", version.Version)
	fmt.Printf(" Branch   : %s\n", version.Branch)
	fmt.Printf(" Revision : %s\n", version.Revision)
	fmt.Printf(" Built    : %s\n", version.Built)
	fmt.Printf("--------------------------------\n\n")
}

func handleSignals(api *api.Torrons) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	<-signalChan
	api.Shutdown()
}
