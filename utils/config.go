package utils

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppUrl                    string        `mapstructure:"APP_URL"`
	AppName                   string        `mapstructure:"APP_NAME"`
	Environment               string        `mapstructure:"ENVIRONMENT"`
	MigrationUrl              string        `mapstructure:"MIGRATION_URL"`
	DBDriver                  string        `mapstructure:"DB_DRIVER"`
	DBSource                  string        `mapstructure:"DB_SOURCE"`
	HTTPAuthServerAddress     string        `mapstructure:"HTTP_AUTH_SERVER_ADDRESS"`
	RedisAddress              string        `mapstructure:"REDIS_ADDRESS"`
	RedisUsername             string        `mapstructure:"REDIS_USERNAME"`
	RedisPwd                  string        `mapstructure:"REDIS_PWD"`
	RefreshTokenSymmetricKey  string        `mapstructure:"REFRESH_TOKEN_SYMMETRIC_KEY"`
	AccessTokenSymmetricKey   string        `mapstructure:"ACCESS_TOKEN_SYMMETRIC_KEY"`
	AccessTokenDuration       time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	RefreshTokenDuration      time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	DBMaxIdleConn             int           `mapstructure:"DB_MAX_IDLE_CONN"`
	DBMaxOpenConn             int           `mapstructure:"DB_MAX_OPEN_CONN"`
	DBMaxIdleTime             int           `mapstructure:"DB_MAX_IDLE_TIME"`
	DBMaxLifeTime             int           `mapstructure:"DB_MAX_LIFE_TIME"`
	SMTPName                  string        `mapstructure:"SMTP_NAME"`
	SMTPAddr                  string        `mapstructure:"SMTP_ADDR"`
	SMTPHost                  string        `mapstructure:"SMTP_HOST"`
	SMTPUsername              string        `mapstructure:"SMTP_USERNAME"`
	SMTPPassword              string        `mapstructure:"SMTP_PASSWORD"`
	GoogleOauthClientId       string        `mapstructure:"GOOGLE_OAUTH_CLIENT_ID"`
	GoogleOauthClientSecret   string        `mapstructure:"GOOGLE_OAUTH_CLIENT_SECRET"`
	GoogleOauthClientRedirect string        `mapstructure:"GOOGLE_REDIRECT"`
}

// LoadConfig reads configuration from a file or environment variables.
// It takes the path to the configuration file as input.
// It returns the loaded configuration and any error encountered during the process.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
