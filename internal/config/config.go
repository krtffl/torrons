package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"

	torrons "github.com/krtffl/torro"
	"github.com/krtffl/torro/internal/logger"
)

const (
	Common string = "common"
	JSON   string = "json"
)

type Logger struct {
	Format string `mapstructure:"format" yaml:"format"`
	Level  string `mapstructure:"level"  yaml:"level"`
	Path   string `mapstructure:"path"   yaml:"path"`
}
type Database struct {
	Host     string `mapstructure:"host"     yaml:"host"`
	Port     uint   `mapstructure:"port"     yaml:"port"`
	User     string `mapstructure:"user"     yaml:"user"`
	Password string `mapstructure:"password" yaml:"password"`
	Name     string `mapstructure:"name"     yaml:"name"`
	SSLMode  string `mapstructure:"ssl"      yaml:"ssl"`
}

type Config struct {
	// Port is the port the HTTP server will listen to
	Port uint `mapstructure:"port" yaml:"port"`

	// Database contains the configuration to connect to the
	// database instance
	Database Database `mapstructure:"database" yaml:"database"`

	// Logger is the logger configuration
	Logger Logger `mapstructure:"logger" yaml:"logger"`
}

// Load loads custom config from specified file or creates a new default one.
// Environment variables take precedence over config file values.
func Load(v *viper.Viper, file string) *Config {
	// Try to load .env file (ignore error if not found)
	if err := godotenv.Load(); err != nil {
		log.Printf("[Config - Load] - No .env file found, using config file and environment variables")
	}

	v.SetConfigFile(file)

	if err := v.ReadInConfig(); err != nil {
		log.Printf(
			"[Config - Load] "+
				"- Couldn't load custom config file. %v. Creating default config...",
			err,
		)

		if err := os.MkdirAll(filepath.Dir(file), 0750); err != nil {
			logger.Fatal("[Config - Load] "+
				"- Couldn't create default config dir. %v", err)
		}

		f, err := os.Create(file)
		if err != nil {
			logger.Fatal("[Config - Load] "+
				"- Couldn't create default cofig file. %v", err)
		}

		defer f.Close()
		if _, err := f.Write(torrons.DefaultConfig); err != nil {
			logger.Fatal("[Config - Load] "+
				"- Couldn't load default config file. %v", err)
		}
	}

	config := &Config{}
	if err := v.Unmarshal(&config); err != nil {
		logger.Fatal("[Config - Load] "+
			"- Couldn't unmarshall config file. %v", err)
	}

	// Override config with environment variables if they exist
	overrideWithEnvVars(config)

	if config.Logger.Format != Common &&
		config.Logger.Format != JSON {
		config.Logger.Format = Common
	}

	initializeLogger(config.Logger)
	return config
}

// overrideWithEnvVars overrides configuration with environment variables
// Environment variables take precedence over config file values
func overrideWithEnvVars(config *Config) {
	// Server configuration
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.ParseUint(port, 10, 32); err == nil {
			config.Port = uint(p)
		}
	}

	// Database configuration
	if host := os.Getenv("DB_HOST"); host != "" {
		config.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		if p, err := strconv.ParseUint(port, 10, 32); err == nil {
			config.Database.Port = uint(p)
		}
	}
	if user := os.Getenv("DB_USER"); user != "" {
		config.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		config.Database.Password = password
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		config.Database.Name = dbName
	}
	if sslMode := os.Getenv("DB_SSL_MODE"); sslMode != "" {
		config.Database.SSLMode = sslMode
	}

	// Logger configuration
	if format := os.Getenv("LOGGER_FORMAT"); format != "" {
		config.Logger.Format = format
	}
	if level := os.Getenv("LOGGER_LEVEL"); level != "" {
		config.Logger.Level = level
	}
	if path := os.Getenv("LOGGER_PATH"); path != "" {
		config.Logger.Path = path
	}
}

func initializeLogger(cfg Logger) {
	logConfg := logger.Configuration{
		EnableConsole: true,
		ConsoleLevel:  logger.GetLevel(cfg.Level),
		FileLevel:     logger.GetLevel(cfg.Level),
		EnableFile:    len(cfg.Path) > 0,
		FileLocation:  cfg.Path,
	}

	if cfg.Format == JSON {
		logConfg.ConsoleJSONFormat = true
		logConfg.FileJSONFormat = true
	}

	logger.NewLogger(logConfg)
}
