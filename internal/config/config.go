package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type WebServerConfig struct {
	RunAddress string `mapstructure:"run-address"` // run address
}

type AppConfig struct {
	Name string `mapstructure:"name"` // app name
}

type LoggingConfiguration struct {
	Level      string `mapstructure:"level"`       // logging level
	Path       string `mapstructure:"path"`        // logging path
	MaxSize    int    `mapstructure:"max-size"`    // logging max size
	MaxBackups int    `mapstructure:"max-backups"` // logging max backups
	MaxAge     int    `mapstructure:"max-age"`     // logging max age
}

type DBConfig struct {
	ConnectionString string `mapstructure:"db-dsn"` // database connection string
}

// Config struct
type Config struct {
	Debug   bool                  `mapstructure:"debug"`   // debug mode
	App     *AppConfig            `mapstructure:"app"`     // app name
	Logging *LoggingConfiguration `mapstructure:"logging"` // logging config
	Web     *WebServerConfig      `mapstructure:"web"`     // web server config
	PostDB  *DBConfig             `mapstructure:"postdb"`  // database config
}

var C *Config = new(Config)

// InitConfiguration returns a new instance of Config
func InitConfiguration() *Config {
	initConfig()
	return C
}

func initConfig() {
	loadDefault()
	loadFile()
	loadEnv()

	if viper.GetBool("debug") {
		viper.SetDefault("logging.level", "debug")
	}

	viper.Unmarshal(C)
}

func loadEnv() {
	viper.SetEnvPrefix("GOAGGREGATOR")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

const (
	AppName = "goaggregator"
)

func loadDefault() {
	viper.SetDefault("debug", false)
	viper.SetDefault("app.name", AppName)

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.path", "logs")
	viper.SetDefault("logging.max-size", 500)
	viper.SetDefault("logging.max-backups", 3)
	viper.SetDefault("logging.max-age", 30)

	viper.SetDefault("web.run-address", ":8100")

	viper.SetDefault("postdb.db-dsn", "postgres://postgres:postgres@localhost:5432/aggregation?sslmode=disable")
}

func loadFile() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AddConfigPath(".")
	viper.AddConfigPath("./config/")
	viper.AddConfigPath("../config/")
	viper.AddConfigPath(fmt.Sprintf("$HOME/.%s", viper.GetString(AppName)))
	viper.AddConfigPath(fmt.Sprintf("/etc/%s/", viper.GetString(AppName)))
	viper.AddConfigPath(fmt.Sprintf("/etc/%s/config/", viper.GetString(AppName)))

	var fileLookupError viper.ConfigFileNotFoundError
	if err := viper.ReadInConfig(); err != nil {
		if errors.As(err, &fileLookupError) {
			if err := viper.WriteConfigAs("./config.yaml"); err != nil {
				fmt.Printf("Error writing config file: %v\n", err)
				panic(err)
			}
		} else {
			fmt.Println(fmt.Printf("Error reading config file, %s. Use default only.\n", err))
		}
	}
}
