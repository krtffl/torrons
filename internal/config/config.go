package config

import (
	"log"
	"os"
	"path/filepath"

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
func Load(v *viper.Viper, file string) *Config {
	v.SetConfigFile(file)

	if err := v.ReadInConfig(); err != nil {
		log.Printf(
			"[Config - Load] "+
				"- Couldn't load custom config file. %v. Creating default config...",
			err,
		)

		if err := os.MkdirAll(filepath.Dir(file), 0770); err != nil {
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

	if config.Logger.Format != Common &&
		config.Logger.Format != JSON {
		config.Logger.Format = Common
	}

	initializeLogger(config.Logger)
	return config
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
