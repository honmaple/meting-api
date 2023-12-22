package app

import (
	"io/ioutil"
	"meting-api/music"
	"os"
	"strings"

	"github.com/honmaple/forest"
	"github.com/honmaple/forest/middleware"
)

type App struct {
	log    *Logger
	Config *Config
	Music  *music.Music
}

func (app *App) Run() error {
	conf := app.Config

	srv := forest.New()
	if conf.GetString("server.mode") == "dev" {
		srv.SetOptions(forest.Debug())
	}

	corsConfig := middleware.CorsConfig{
		AllowOrigins: conf.GetStringSlice("server.cors.allow_origins"),
		AllowMethods: conf.GetStringSlice("server.cors.allow_methods"),
		AllowHeaders: conf.GetStringSlice("server.cors.allow_headers"),
	}
	if len(corsConfig.AllowOrigins) == 0 {
		corsConfig.AllowOrigins = middleware.DefaultCorsConfig.AllowOrigins
	}
	if len(corsConfig.AllowMethods) == 0 {
		corsConfig.AllowMethods = middleware.DefaultCorsConfig.AllowMethods
	}
	if len(corsConfig.AllowHeaders) == 0 {
		corsConfig.AllowHeaders = middleware.DefaultCorsConfig.AllowHeaders
	}

	srv.Use(middleware.Logger(), middleware.CorsWithConfig(corsConfig))
	srv.GET("/meting", app.aplayer)
	return srv.Start(conf.GetString("server.addr"))
}

func (app *App) Init(file string, strs ...string) error {
	conf := app.Config

	if _, err := os.Stat(file); err == nil || os.IsExist(err) {
		content, err := ioutil.ReadFile(file)
		if err != nil {
			return err
		}
		conf.SetConfigFile(file)
		if err := conf.ReadConfig(strings.NewReader(os.ExpandEnv(string(content)))); err != nil {
			return err
		}
	}

	for _, str := range strs {
		c := strings.SplitN(str, "=", 2)
		if len(c) < 2 {
			conf.Set(c[0], "")
		} else {
			conf.Set(c[0], c[1])
		}
	}

	log, err := NewLogger(conf)
	if err != nil {
		return err
	}
	app.log = log
	app.Music = music.New(app.Config.Viper)
	return nil
}

func New() *App {
	return &App{
		Config: DefaultConfig(),
	}
}
