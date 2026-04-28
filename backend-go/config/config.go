package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port string
	Mode string
}

type DatabaseConfig struct {
	Type     string
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	Secret     string
	ExpireTime int64
}

var config *Config

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.dbname", "parking.db")
	viper.SetDefault("jwt.secret", "parking-system-jwt-secret-key")
	viper.SetDefault("jwt.expireTime", 86400)

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Config file not found, using defaults: %v", err)
	}

	config = &Config{
		Server: ServerConfig{
			Port: viper.GetString("server.port"),
			Mode: viper.GetString("server.mode"),
		},
		Database: DatabaseConfig{
			Type:     viper.GetString("database.type"),
			Host:     viper.GetString("database.host"),
			Port:     viper.GetString("database.port"),
			User:     viper.GetString("database.user"),
			Password: viper.GetString("database.password"),
			DBName:   viper.GetString("database.dbname"),
			SSLMode:  viper.GetString("database.sslmode"),
		},
		JWT: JWTConfig{
			Secret:     viper.GetString("jwt.secret"),
			ExpireTime: viper.GetInt64("jwt.expireTime"),
		},
	}
}

func GetConfig() *Config {
	return config
}
