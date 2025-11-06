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
	"github.com/krtffl/torro/internal/migrations"
	"github.com/krtffl/torro/version"
)

var (
	configPath = flag.String(
		"config",
		"config/config.yaml",
		"path from where the config file will be loaded",
	)
	migrationsPath = flag.String(
		"migrations",
		"migrations",
		"path to migrations directory",
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

	// Run migrations automatically unless skipped
	if !*skipMigrations {
		if err := runMigrations(cfg, *migrationsPath); err != nil {
			logger.Error("Migration failed: %v", err)
			logger.Warn("Starting server anyway. Run 'make migrate' to fix migrations.")
		} else {
			logger.Info("Migrations completed successfully")
		}
	} else {
		logger.Info("Skipping automatic migrations (--skip-migrations flag set)")
	}

	api := api.New(cfg)

	go api.Run()
	handleSignals(api)
}

func runMigrations(cfg *config.Config, migrationsPath string) error {
	// Build database URL from config
	dbURL := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	logger.Info("Running database migrations from: %s", migrationsPath)

	// Check current version
	version, dirty, err := migrations.Version(dbURL, migrationsPath)
	if err != nil {
		logger.Warn("Could not get migration version: %v", err)
	} else if dirty {
		logger.Warn("Database is in dirty state at version %d. You may need to run 'make migrate-force'", version)
	} else if version > 0 {
		logger.Info("Current migration version: %d", version)
	}

	// Run migrations
	if err := migrations.Run(dbURL, migrationsPath); err != nil {
		return err
	}

	return nil
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
