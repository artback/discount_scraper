package config

import (
	"errors"
	"fmt"
	"grocery_scraper/internal/models"
	"log"

	"github.com/fsnotify/fsnotify"

	"github.com/spf13/viper"
)

// Config holds the application configuration parameters.
type Config struct {
	DBConn   string
	Stores   []models.Store
	AIAPIKey string
}

// Global constants for configuration keys
const (
	DBHostKey     = "DB_HOST"
	DBPortKey     = "DB_PORT"
	DBUserKey     = "DB_USER"
	DBPasswordKey = "DB_PASSWORD"
	DBNameKey     = "DB_NAME"
	StoresKey     = "stores" // Key for the list of stores in config.yaml
	AIAPIKey      = "AI_API_KEY"
)

// Init initializes Viper, sets defaults, and constructs the DSN.
func Init() *Config {
	// --- File-based configuration ---
	viper.SetConfigName("config") // name of config file (e.g., config.yaml)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".") // look in the current directory

	// Attempt to read the config file
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			// Config file not found; this is not an error, we can rely on defaults/env
			log.Println("config.yaml not found, using default stores and environment variables.")
		}
	}

	// Set up Viper to read environment variables
	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()

	// Construct the DSN from individual components
	dsn := buildDSN()
	// Unmarshal the stores configuration
	var stores []models.Store
	if err := viper.UnmarshalKey(StoresKey, &stores); err != nil {
		log.Fatalf("Fatal Error: could not unmarshal stores configuration: %v", err)
	}
	viper.OnConfigChange(func(e fsnotify.Event) {
	})

	viper.WatchConfig()

	return &Config{
		DBConn:   dsn,
		Stores:   stores,
		AIAPIKey: viper.GetString(AIAPIKey),
	}
}

// buildDSN constructs the PostgreSQL DSN from individual config values read by Viper.
func buildDSN() string {
	host := viper.GetString(DBHostKey)
	port := viper.GetString(DBPortKey)
	user := viper.GetString(DBUserKey)
	password := viper.GetString(DBPasswordKey)
	dbname := viper.GetString(DBNameKey)

	if host == "" || user == "" || dbname == "" {
		log.Fatalf("Fatal Error: Missing mandatory database configuration (Host: %s, User: %s, DB Name: %s). Check ENV variables or config file.", host, user, dbname)
	}

	// Standard PostgreSQL DSN format
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Stockholm",
		host, user, password, dbname, port,
	)
	return dsn
}
