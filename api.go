package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/honmaple/forest"
	"github.com/honmaple/forest/middleware"
	"github.com/tidwall/gjson"
	"github.com/urfave/cli/v2"
)

type (
	Config struct {
		Addr       string `json:"addr"`
		Host       string `json:"host"`
		NeteaseApi string `json:"netease_api"`
	}
	Meting struct {
		config Config
	}
	Result struct {
		Title  string `json:"title"`
		Author string `json:"author"`
		Lrc    string `json:"lrc"`
		Pic    string `json:"pic"`
		Url    string `json:"url"`
	}
)

func (m Meting) song(c forest.Context) error {
	id := c.QueryParam("id")
	if id == "" {
		return c.JSON(400, map[string]interface{}{"msg": "参数错误"})
	}

	resp, err := http.Get(fmt.Sprintf("%s/song/url?id=%s", m.config.NeteaseApi, id))
	if err != nil {
		c.Logger().Errorln(err.Error())
		return forest.ErrInternalServerError
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.Logger().Errorln(err.Error())
		return forest.ErrInternalServerError
	}
	urls := gjson.GetBytes(buf, "data").Array()
	if len(urls) == 0 {
		return forest.ErrNotFound
	}
	return c.Redirect(302, urls[0].Get("url").String())
}

func (m Meting) lyric(c forest.Context) error {
	id := c.QueryParam("id")
	if id == "" {
		return c.JSON(400, map[string]interface{}{"msg": "参数错误"})
	}
	resp, err := http.Get(fmt.Sprintf("%s/lyric?id=%s", m.config.NeteaseApi, id))
	if err != nil {
		c.Logger().Errorln(err.Error())
		return forest.ErrInternalServerError
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.Logger().Errorln(err.Error())
		return forest.ErrInternalServerError
	}
	return c.String(200, gjson.GetBytes(buf, "lrc.lyric").String())
}

func (m Meting) aplayer(c forest.Context) error {
	id := c.QueryParam("id")
	if id == "" {
		return c.JSON(400, map[string]interface{}{"msg": "参数错误"})
	}

	var (
		err  error
		resp *http.Response
	)

	typ := c.QueryParam("type")
	switch typ {
	case "playlist":
		resp, err = http.Get(fmt.Sprintf("%s/playlist/track/all?id=%s", m.config.NeteaseApi, id))
	case "artist":
		resp, err = http.Get(fmt.Sprintf("%s/artist/top/song?id=%s", m.config.NeteaseApi, id))
	case "song":
		resp, err = http.Get(fmt.Sprintf("%s/song/detail?ids=%s", m.config.NeteaseApi, id))
	default:
		return forest.ErrNotFound
	}

	if err != nil {
		c.Logger().Errorln(err.Error())
		return forest.ErrInternalServerError
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.Logger().Errorln(err.Error())
		return forest.ErrInternalServerError
	}
	songs := gjson.GetBytes(buf, "songs").Array()
	results := make([]Result, len(songs))
	for i, song := range songs {
		songId := song.Get("id").Int()
		ars := song.Get("ar").Array()

		authors := make([]string, len(ars))
		for i, ar := range ars {
			authors[i] = ar.Get("name").String()
		}
		results[i] = Result{
			Title:  song.Get("name").String(),
			Author: strings.Join(authors, ", "),
			Pic:    song.Get("al.picUrl").String(),
			Url:    fmt.Sprintf("%s/aplayer/song?id=%d", m.config.Host, songId),
			Lrc:    fmt.Sprintf("%s/aplayer/lyric?id=%d", m.config.Host, songId),
		}
	}
	return c.JSON(200, results)
}

func action(clx *cli.Context) error {
	r := forest.New(forest.Debug())
	r.Use(middleware.Logger())
	r.Use(middleware.Cors())

	config := Config{
		Addr:       ":8000",
		Host:       "http://localhost:8000",
		NeteaseApi: "http://localhost:3000",
	}
	if api := clx.String("api"); api != "" {
		config.NeteaseApi = api
	}
	if addr := clx.String("addr"); addr != "" {
		config.Addr = addr
	}
	if host := clx.String("host"); host != "" {
		config.Host = host
	}

	meting := Meting{
		config: config,
	}
	r.GET("/aplayer/song", meting.song)
	r.GET("/aplayer/lyric", meting.lyric)
	r.GET("/aplayer", meting.aplayer)

	return r.Start(config.Addr)
}

func main() {
	app := &cli.App{
		Name:  "meting-api",
		Usage: "meting api",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "api",
				Usage: "netease api",
			},
			&cli.StringFlag{
				Name:    "addr",
				Aliases: []string{"a"},
				Usage:   "listen addr",
			},
			&cli.StringFlag{
				Name:  "host",
				Usage: "server domain",
			},
		},
		Action: action,
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err.Error())
	}
}
