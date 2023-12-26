package app

import (
	"embed"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"meting-api/music"

	"github.com/honmaple/forest"
	"github.com/honmaple/forest/middleware"
	"github.com/honmaple/forest/render"
	"github.com/nutsdb/nutsdb"
)

type App struct {
	Log    *Logger
	Cache  *nutsdb.DB
	Music  *music.Music
	Config *Config
}

var (
	//go:embed templates/*
	templateFS embed.FS
)

func (app *App) ShowConfig() {
	keys := app.Config.AllKeys()
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Println(k, "=", app.Config.Get(k))
	}
}

func (app *App) Set(key string, value any) {
	app.Config.Set(key, value)
}

func (app *App) Run() error {
	defer app.Close()

	conf := app.Config

	srv := forest.New()
	if conf.GetString("server.mode") == "dev" {
		srv.SetOptions(forest.Debug())
	}
	srv.Renderer = render.NewTemplate(templateFS, "templates/*.html")

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
	srv.Use(app.cacheResponse)
	srv.GET("/", app.meting)
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
	if err := app.initCache(); err != nil {
		return err
	}
	log, err := NewLogger(conf)
	if err != nil {
		return err
	}
	app.Log = log
	app.Music = music.New(app.Config.Viper)
	return nil
}

func (app *App) Close() error {
	if app.Cache != nil {
		return app.Cache.Close()
	}
	return nil
}

func New() *App {
	return &App{
		Config: DefaultConfig(),
	}
}
