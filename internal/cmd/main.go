package cmd

import (
	"embed"
	"fmt"
	"os"
	"strings"

	"meting-api/internal/app"
	"meting-api/music"

	"github.com/urfave/cli/v2"
)

const (
	PROCESS     = "meting-api"
	DESCRIPTION = "meting api"
	VERSION     = "0.1.0"
)

var (
	FS         embed.FS
	defaultApp = app.New()
)

func before(clx *cli.Context) error {
	return defaultApp.Init(clx.String("config"), clx.StringSlice("set-config")...)
}

func action(clx *cli.Context) error {
	if clx.Bool("list") {
		fmt.Println(strings.Join(music.List(), ", "))
		return nil
	}
	if addr := clx.String("addr"); addr != "" {
		defaultApp.Set("server.addr", addr)
	}
	if debug := clx.Bool("debug"); debug {
		defaultApp.Set("server.mode", "dev")
	}
	return defaultApp.Run()
}

func Run() {
	app := &cli.App{
		Name:    PROCESS,
		Usage:   DESCRIPTION,
		Version: VERSION,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"D"},
				Usage:   "debug mode",
			},
			&cli.StringFlag{
				Name:    "addr",
				Aliases: []string{"a"},
				Usage:   "listen `ADDR`",
			},
			&cli.BoolFlag{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "list music servers",
			},
			&cli.PathFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "load config from `FILE`",
				Value:   "config.yaml",
			},
			&cli.StringSliceFlag{
				Name:  "set-config",
				Usage: "set config from string",
			},
		},
		Before: before,
		Action: action,
		Commands: []*cli.Command{
			&cli.Command{
				Name:  "cache-delete",
				Usage: "delete cache from key",
				Action: func(clx *cli.Context) error {
					return defaultApp.DeleteCache(clx.Args().First())
				},
			},
			&cli.Command{
				Name:  "config",
				Usage: "show all config",
				Action: func(clx *cli.Context) error {
					defaultApp.ShowConfig()
					return nil
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err.Error())
	}
}
