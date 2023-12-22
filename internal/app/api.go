package app

import (
	"fmt"
	"meting-api/music"

	"github.com/honmaple/forest"
)

type (
	Result struct {
		Title  string `json:"title"`
		Author string `json:"author"`
		Lrc    string `json:"lrc"`
		Pic    string `json:"pic"`
		Url    string `json:"url"`
	}
)

func (app *App) result(c forest.Context, server string, song *music.Song) *Result {
	return &Result{
		Title:  song.Name,
		Author: song.GetArtist(),
		Pic:    song.Picture,
		Lrc:    fmt.Sprintf("%s/meting?server=%s&type=lrc&id=%s", c.Request().Host, server, song.Id),
		Url:    fmt.Sprintf("%s/meting?server=%s&type=url&id=%s", c.Request().Host, server, song.Id),
	}
}

func (app *App) aplayer(c forest.Context) error {
	id := c.QueryParam("id")
	if id == "" {
		return c.JSON(400, map[string]interface{}{"msg": "参数错误123"})
	}

	server := c.QueryParam("server")
	if server == "" {
		server = "netease"
	}
	switch c.QueryParam("type") {
	case "name":
		data, err := app.Music.Search(server, id)
		if err != nil {
			return c.String(400, err.Error())
		}
		results := make([]*Result, len(data))
		for i, song := range data {
			results[i] = app.result(c, server, song)
		}
		return c.JSON(200, results)
	case "song":
		data, err := app.Music.Song(server, id)
		if err != nil {
			return c.String(400, err.Error())
		}
		results := []*Result{
			app.result(c, server, data),
		}
		return c.JSON(200, results)
	case "url":
		data, err := app.Music.SongLink(server, id)
		if err != nil {
			return c.String(400, err.Error())
		}
		return c.Redirect(302, data.URL)
	case "pic":
		data, err := app.Music.Song(server, id)
		if err != nil {
			return c.String(400, err.Error())
		}
		return c.Redirect(302, data.Picture)
	case "lrc":
		data, err := app.Music.Lyric(server, id)
		if err != nil {
			return c.String(400, err.Error())
		}
		return c.String(200, data.Lyric)
	case "album":
		data, err := app.Music.Album(server, id)
		if err != nil {
			return c.String(400, err.Error())
		}
		results := make([]*Result, len(data.Songs))
		for i, song := range data.Songs {
			results[i] = app.result(c, server, song)
		}
		return c.JSON(200, results)
	case "artist":
		data, err := app.Music.Artist(server, id)
		if err != nil {
			return c.String(400, err.Error())
		}
		results := make([]*Result, len(data.Songs))
		for i, song := range data.Songs {
			results[i] = app.result(c, server, song)
		}
		return c.JSON(200, results)
	case "playlist":
		data, err := app.Music.Playlist(server, id)
		if err != nil {
			return c.String(400, err.Error())
		}
		results := make([]*Result, len(data.Songs))
		for i, song := range data.Songs {
			results[i] = app.result(c, server, song)
		}
		return c.JSON(200, results)
	default:
		return c.JSON(400, map[string]interface{}{"msg": "参数错误321"})
	}
}
