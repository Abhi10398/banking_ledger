package config

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type Configuration struct {
	Port           int64            `validate:"required"`
	LogLevel       string           `validate:"required"`
	PostgresConfig PostgresDbConfig `validate:"required"`
	PostgresDbName string           `validate:"required"`
	JwtSecretKey   string           `validate:"required"`
	ClientConfig   ClientConfig     `validate:"required"`
}

type ClientConfig struct {
	CMSClient struct {
		BaseURL       string `validate:"required"`
		ClientId      string `validate:"required"`
		Authorization string `validate:"required"`
		HTTPConnectionPool
	}
}
type PostgresDbConfig struct {
	Host            string
	Port            int64
	Name            string
	User            string
	Password        string
	MaxIdleConn     int64
	ConnMaxIdleTime int64
	ConnMaxLifetime int64
	MaxOpenConn     int64
}

type HTTPConnectionPool struct {
	TimeoutInMs               time.Duration
	RetryCount                int
	RetryWaitTimeInMs         time.Duration
	MaxIdleConnections        int
	MaxIdleConnectionsPerHost int
	IdleConnTimeoutInSec      time.Duration
}

var (
	config     *Configuration
	configOnce sync.Once
)

func Load() *Configuration {
	configOnce.Do(func() {
		v := viper.New()
		v.AutomaticEnv()
		vaultEnabled := v.GetString("VAULT_ENABLED")

		v.SetConfigName("static")
		if strings.EqualFold(vaultEnabled, "TRUE") {
			v.AddConfigPath("./")
			v.AddConfigPath("../")
			v.AddConfigPath("/vault/secrets/")
			v.SetConfigType("json")
		} else {
			v.AddConfigPath("./config")
			v.AddConfigPath("../config")
			v.AddConfigPath("../../config")
			v.AddConfigPath("../../../config")
			v.SetConfigType("yaml")
		}

		// Read the static config
		if err := v.ReadInConfig(); err != nil {
			panic(fmt.Errorf("error while reading config file, error - %v", err))
		}
		if err := v.Unmarshal(&config); err != nil {
			panic(fmt.Errorf("error while unmarshalling config file, error - %v", err))
		}
	})
	return config
}

func PostgresConf() PostgresDbConfig {
	return config.PostgresConfig
}

func PostgresDbName() string {
	return config.PostgresDbName
}

func GetJwtSecretKey() string {
	return config.JwtSecretKey
}

func GetClientConf() ClientConfig {
	return config.ClientConfig
}
