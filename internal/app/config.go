package app

import (
	"strings"

	"github.com/spf13/viper"
)

var (
	defaultConfig = map[string]interface{}{
		"cache.ttl":                  3600,
		"cache.path":                 "./cache",
		"cache.enabled":              true,
		"server.addr":                ":8000",
		"server.host":                "",
		"server.mode":                "",
		"server.cors.allow_headers":  []string{"*"},
		"server.cors.allow_origins":  []string{"*"},
		"server.cors.allow_methods":  nil,
		"logger.out":                 "stdout",
		"logger.level":               "info",
		"logger.file_format":         "%Y%m%d",
		"logger.file_rotation_count": 3,
		"netease_api.host":           "",
	}
)

type Config struct {
	*viper.Viper
}

func DefaultConfig() *Config {
	conf := &Config{
		Viper: viper.New(),
	}
	for k, v := range defaultConfig {
		conf.SetDefault(k, v)
	}

	// METING_SERVER_ADDR=:8001 METING_SERVER_HOST=https://api.example.com/meting meting-api -D
	conf.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	conf.SetEnvPrefix("meting")
	conf.AutomaticEnv()
	return conf
}
