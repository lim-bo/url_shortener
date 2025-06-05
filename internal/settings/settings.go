package settings

import (
	"log"
	"sync"

	"github.com/spf13/viper"
)

type Config struct {
}

func (c *Config) GetString(key string) string {
	return viper.GetString(key)
}

var (
	instance *Config
	once     sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		viper.AddConfigPath("config")
		viper.SetConfigType("yaml")
		viper.SetConfigName("cfg")
		err := viper.ReadInConfig()
		if err != nil {
			log.Fatal("reading config error: ", err)
		}
		instance = &Config{}
	})
	return instance
}
