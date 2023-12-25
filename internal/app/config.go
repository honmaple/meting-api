package app

import (
	"github.com/spf13/viper"
)

var (
	defaultConfig = map[string]interface{}{
		"cache.ttl":                  3600,
		"cache.path":                 "./cache",
		"server.addr":                ":8000",
		"server.host":                "",
		"server.mode":                "",
		"logger.out":                 "stdout",
		"logger.level":               "info",
		"logger.file_format":         "%Y%m%d",
		"logger.file_rotation_count": 3,
	}
)

type Config struct {
	*viper.Viper
}

func DefaultConfig() *Config {
	conf := &Config{
		Viper: viper.New(),
	}
	configs := []map[string]interface{}{
		defaultConfig,
	}
	for _, config := range configs {
		for k, v := range config {
			conf.SetDefault(k, v)
		}
	}
	return conf
}
