package app

import (
	"fmt"
	"strings"

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

func (app *App) getHost(c forest.Context) string {
	host := app.Config.GetString("server.host")
	if host == "" {
		host = c.Request().Host
	}
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}
	return host
}

func (app *App) toResult(c forest.Context, server string, song *music.Song) *Result {
	host := app.getHost(c)
	return &Result{
		Title:  song.Name,
		Author: song.GetArtist(),
		Pic:    song.Picture,
		Lrc:    fmt.Sprintf("%s?server=%s&type=lrc&id=%s", host, server, song.Id),
		Url:    fmt.Sprintf("%s?server=%s&type=url&id=%s", host, server, song.Id),
	}
}

func (app *App) meting(c forest.Context) error {
	id := c.QueryParam("id")
	if id == "" {
		return c.Render(200, "index.html", forest.H{"host": app.getHost(c), "msg": "id不能为空"})
	}

	server := c.QueryParam("server", "netease")

	switch c.QueryParam("type") {
	case "name":
		data, err := app.Music.Search(server, id)
		if err != nil {
			return c.JSON(400, forest.H{"msg": err.Error()})
		}
		results := make([]*Result, len(data))
		for i, song := range data {
			results[i] = app.toResult(c, server, song)
		}
		return c.JSON(200, results)
	case "song":
		data, err := app.Music.Song(server, id)
		if err != nil {
			return c.JSON(400, forest.H{"msg": err.Error()})
		}
		results := []*Result{
			app.toResult(c, server, data),
		}
		return c.JSON(200, results)
	case "url":
		data, err := app.Music.SongLink(server, id)
		if err != nil {
			return c.JSON(400, forest.H{"msg": err.Error()})
		}
		return c.Redirect(302, data.URL)
	case "pic":
		data, err := app.Music.Song(server, id)
		if err != nil {
			return c.JSON(400, forest.H{"msg": err.Error()})
		}
		return c.Redirect(302, data.Picture)
	case "lrc":
		data, err := app.Music.Lyric(server, id)
		if err != nil {
			return c.JSON(400, forest.H{"msg": err.Error()})
		}
		return c.String(200, data.Lyric)
	case "album":
		data, err := app.Music.Album(server, id)
		if err != nil {
			return c.JSON(400, forest.H{"msg": err.Error()})
		}
		results := make([]*Result, len(data.Songs))
		for i, song := range data.Songs {
			results[i] = app.toResult(c, server, song)
		}
		return c.JSON(200, results)
	case "artist":
		data, err := app.Music.Artist(server, id)
		if err != nil {
			return c.JSON(400, forest.H{"msg": err.Error()})
		}
		results := make([]*Result, len(data.Songs))
		for i, song := range data.Songs {
			results[i] = app.toResult(c, server, song)
		}
		return c.JSON(200, results)
	case "playlist":
		data, err := app.Music.Playlist(server, id)
		if err != nil {
			return c.JSON(400, forest.H{"msg": err.Error()})
		}
		results := make([]*Result, len(data.Songs))
		for i, song := range data.Songs {
			results[i] = app.toResult(c, server, song)
		}
		return c.JSON(200, results)
	default:
		return c.Render(200, "index.html", forest.H{"host": app.getHost(c), "msg": "type错误"})
	}
}
