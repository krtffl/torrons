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
	"github.com/krtffl/torro/internal/logger"
	"github.com/krtffl/torro/version"
)

var (
	configPath = flag.String(
		"config",
		"config/config.yaml",
		"path from where the config file will be loaded",
	)
	skipMigrations = flag.Bool(
		"skip-migrations",
		false,
		"skip automatic database migrations on startup",
	)
)

func main() {
	flag.Parse()
	banner()

	cfg := config.Load(viper.New(), *configPath)

	// Migrations run once, inside api.New -> NewDatabaseConnection, against the
	// embedded migration files, gated by this flag. There used to be a second,
	// file-based migration pass here that duplicated that work and failed on
	// container images that don't ship the migrations directory, logging a
	// misleading "Migration failed" error on every boot.
	if *skipMigrations {
		logger.Info("Skipping automatic migrations (--skip-migrations flag set)")
	}

	api := api.New(cfg, !*skipMigrations)

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
