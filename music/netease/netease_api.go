package netease

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"meting-api/music"

	"github.com/tidwall/gjson"
)

type neteaseApi struct {
	client *http.Client
	config *music.Config
}

func (self *neteaseApi) request(method, url string, data map[string]string) ([]byte, error) {
	host := self.config.GetString("netease_api.host")
	if host == "" {
		return nil, errors.New("netease api is required")
	}
	url = strings.TrimSuffix(host, "/") + url

	var body io.Reader

	if method != "GET" && data != nil {
		body = fromData(data)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if method == "GET" && data != nil {
		q := req.URL.Query()
		for k, v := range data {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()
	}

	headers := map[string]string{
		"Referer":         "https://music.163.com",
		"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_5, AppleWebKit/605.1.15 (KHTML, like Gecko,",
		"X-Real-IP":       randomIP(),
		"Accept":          "*/*",
		"Accept-Language": "zh-CN,zh;q=0.8,gl;q=0.6,zh-TW;q=0.4",
		"Connection":      "keep-alive",
		"Content-Type":    "application/x-www-form-urlencoded",
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for k, v := range self.config.GetStringMapString("netease_api.headers") {
		req.Header.Set(k, v)
	}

	resp, err := self.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("result is null")
	}

	return ioutil.ReadAll(resp.Body)
}

func (self *neteaseApi) toSong(result gjson.Result) *music.Song {
	song := &music.Song{
		Id:      result.Get("id").String(),
		Name:    result.Get("name").String(),
		Picture: result.Get("al.picUrl").String(),
	}
	song.Picture = strings.ReplaceAll(song.Picture, "http://", "https://")
	if arts := result.Get("ar").Array(); len(arts) > 0 {
		artist := make([]*music.Artist, len(arts))
		for i, art := range arts {
			artist[i] = &music.Artist{
				Id:   art.Get("id").String(),
				Name: art.Get("name").String(),
			}
		}
		song.Artist = artist
	}
	return song
}

func (self *neteaseApi) Song(id string) (*music.Song, error) {
	data := map[string]string{
		"ids": id,
	}

	res, err := self.request("GET", "/song/detail", data)
	if err != nil {
		return nil, err
	}
	result := gjson.ParseBytes(res).Get("songs.0")
	return self.toSong(result), nil
}

func (self *neteaseApi) SongLink(id string) (*music.SongLink, error) {
	data := map[string]string{
		"id": id,
		// "br": fmt.Sprintf("%d", 320*1000),
	}

	res, err := self.request("GET", "/song/url", data)
	if err != nil {
		return nil, err
	}

	result := gjson.ParseBytes(res)

	link := &music.SongLink{
		Id:   result.Get("data.0.id").String(),
		URL:  result.Get("data.0.url").String(),
		Br:   result.Get("data.0.br").Int(),
		Size: result.Get("data.0.size").Int(),
	}
	link.URL = strings.ReplaceAll(link.URL, "http://", "https://")
	return link, nil
}

func (self *neteaseApi) Album(id string) (*music.Album, error) {
	data := map[string]string{
		"id": id,
	}

	res, err := self.request("GET", "/album", data)
	if err != nil {
		return nil, err
	}

	result := gjson.ParseBytes(res)

	album := &music.Album{
		Id:   result.Get("album.id").String(),
		Name: result.Get("album.name").String(),
	}
	if sgs := result.Get("songs").Array(); len(sgs) > 0 {
		songs := make([]*music.Song, len(sgs))
		for i, sg := range sgs {
			songs[i] = self.toSong(sg)
		}
		album.Songs = songs
	}
	return album, nil
}

func (self *neteaseApi) Lyric(id string) (*music.Lyric, error) {
	data := map[string]string{
		"id": id,
	}

	res, err := self.request("GET", "/lyric", data)
	if err != nil {
		return nil, err
	}

	result := gjson.ParseBytes(res)

	lrc := &music.Lyric{
		Lyric: result.Get("lrc.lyric").String(),
		Trans: result.Get("tlyric.lyric").String(),
	}
	return lrc, nil
}

func (self *neteaseApi) Artist(id string) (*music.Artist, error) {
	data := map[string]string{
		"id":    id,
		"order": "hot",
		"limit": "50",
	}

	res, err := self.request("GET", "/artist/songs", data)
	if err != nil {
		return nil, err
	}

	result := gjson.ParseBytes(res)

	ins := &music.Artist{
		Id: id,
	}
	if hots := result.Get("songs").Array(); len(hots) > 0 {
		songs := make([]*music.Song, len(hots))
		for i, hot := range hots {
			songs[i] = self.toSong(hot)
		}
		ins.Songs = songs
	}
	return ins, nil
}

func (self *neteaseApi) Playlist(id string) (*music.Playlist, error) {
	data := map[string]string{
		"id": id,
	}

	res, err := self.request("GET", "/playlist/track/all", data)
	if err != nil {
		return nil, err
	}

	result := gjson.ParseBytes(res)

	ins := &music.Playlist{
		Id: id,
	}
	if tracks := result.Get("songs").Array(); len(tracks) > 0 {
		songs := make([]*music.Song, len(tracks))
		for i, track := range tracks {
			songs[i] = self.toSong(track)
		}
		ins.Songs = songs
	}
	return ins, nil
}

func (self *neteaseApi) Search(keyword string) ([]*music.Song, error) {
	data := map[string]string{
		"type":     "1",
		"limit":    "30",
		"keywords": keyword,
	}

	res, err := self.request("GET", "/cloudsearch", data)
	if err != nil {
		return nil, err
	}

	result := gjson.ParseBytes(res)

	songs := make([]*music.Song, 0)
	for _, song := range result.Get("result.songs").Array() {
		songs = append(songs, self.toSong(song))
	}
	return songs, nil
}

func init() {
	music.Register("netease_api", func(config *music.Config, client *http.Client) music.Server {
		return &neteaseApi{
			config: config,
			client: client,
		}
	})
}
