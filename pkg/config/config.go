package config

import (
	"log"
	"sync"

	"github.com/spf13/viper"
)

var once sync.Once

var AppConfig *Config

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Jwt       JwtConfig       `mapstructure:"jwt"`
	Snowflake SnowflakeConfig `mapstructure:"snowflake"`
	Log       LogConfig       `mapstructure:"log"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Driver       string `mapstructure:"driver"`
	DSN          string `mapstructure:"dsn"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JwtConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

type SnowflakeConfig struct {
	NodeID int `mapstructure:"node_id"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
}

func InitConfig() {
	once.Do(func() {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./config")

		// 设置环境变量前缀
		viper.SetEnvPrefix("INSURAI")
		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("Error reading config file: %v", err)
		}

		if err := viper.Unmarshal(&AppConfig); err != nil {
			log.Fatalf("Error unmarshalling config: %v", err)
		}
		log.Printf("Config: %v", AppConfig)
	})
}
