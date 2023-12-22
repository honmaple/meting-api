package tencent

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"meting-api/music"

	"github.com/tidwall/gjson"
)

type tencent struct {
	client *http.Client
	config *music.Config
}

func (self *tencent) request(method, url string, data map[string]string) ([]byte, error) {
	var body io.Reader

	if method != "GET" && data != nil {
		body = formData(data)
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
		"Referer":         "http://y.qq.com",
		"Cookie":          "pgv_pvi=22038528; pgv_si=s3156287488; pgv_pvid=5535248600; yplayer_open=1; ts_last=y.qq.com/portal/player.html; ts_uid=4847550686; yq_index=0; qqmusic_fromtag=66; player_exist=1",
		"User-Agent":      "QQ%E9%9F%B3%E4%B9%90/54409 CFNetwork/901.1 Darwin/17.6.0 (x86_64)",
		"Accept":          "*/*",
		"Accept-Language": "zh-CN,zh;q=0.8,gl;q=0.6,zh-TW;q=0.4",
		"Connection":      "keep-alive",
		"Content-Type":    "application/x-www-form-urlencoded",
	}
	for k, v := range headers {
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

func (self *tencent) toSong(result gjson.Result) *music.Song {
	song := &music.Song{
		Id:      result.Get("id").String(),
		Name:    result.Get("name").String(),
		Picture: fmt.Sprintf("https://y.gtimg.cn/music/photo_new/T002R300x300M000%s.jpg", result.Get("album.mid").String()),
	}
	if arts := result.Get("singer").Array(); len(arts) > 0 {
		artist := make([]*music.Artist, len(arts))
		for i, art := range arts {
			artist[i] = &music.Artist{
				Id:   art.Get("mid").String(),
				Name: art.Get("name").String(),
			}
		}
		song.Artist = artist
	}
	return song
}

func (self *tencent) Song(id string) (*music.Song, error) {
	data := map[string]string{
		"songmid":  id,
		"platform": "yqq",
		"format":   "json",
	}

	res, err := self.request("GET", "https://c.y.qq.com/v8/fcg-bin/fcg_play_single_song.fcg", data)
	if err != nil {
		return nil, err
	}
	result := gjson.ParseBytes(res).Get("data.0")
	return self.toSong(result), nil
}

func (self *tencent) SongLink(id string) (*music.SongLink, error) {
	uin := "0"
	guid := string(randomBytes(7, "1234567890"))

	// 'size_flac', 999, 'F000', 'flac'
	// 'size_320mp3', 320, 'M800', 'mp3'
	// 'size_192aac', 192, 'C600', 'm4a'
	// 'size_128mp3', 128, 'M500', 'mp3'
	// 'size_96aac', 96, 'C400', 'm4a'
	// 'size_48aac', 48, 'C200', 'm4a'
	// 'size_24aac', 24, 'C100', 'm4a'
	d := map[string]any{
		"req_0": map[string]any{
			"module": "vkey.GetVkeyServer",
			"method": "CgiGetVkey",
			"param": map[string]any{
				"guid":      guid,
				"uin":       uin,
				"loginflag": 1,
				"platform":  "20",
				"filename": []string{
					// fmt.Sprintf("F000%s.flac", id),
					// fmt.Sprintf("M800%s.mp3", id),
					fmt.Sprintf("M500%s.mp3", id),
				},
				"songmid":  []string{id},
				"songtype": []int{0},
			},
		},
	}
	b, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}

	data := map[string]string{
		"data":        string(b),
		"format":      "json",
		"platform":    "yqq.json",
		"needNewCode": "0",
	}

	res, err := self.request("GET", "https://u.y.qq.com/cgi-bin/musicu.fcg", data)
	if err != nil {
		return nil, err
	}

	result := gjson.ParseBytes(res)

	domain := ""
	for _, d := range result.Get("req_0.data.sip").Array() {
		if sip := d.String(); !strings.HasPrefix(sip, "http://ws.") {
			domain = sip
			break
		}
	}
	url := result.Get("req_0.data.midurlinfo.0.purl").String()
	if domain == "" || url == "" {
		return nil, errors.New("no url found")
	}

	link := &music.SongLink{
		Id:  result.Get("req_0.data.midurlinfo.0.songmid").String(),
		URL: fmt.Sprintf("%s%s", strings.ReplaceAll(domain, "http://", "https://"), url),
		Br:  128,
	}
	return link, nil
}

func (self *tencent) Album(id string) (*music.Album, error) {
	data := map[string]string{
		"albummid": id,
		"format":   "json",
		"platform": "yqq",
		"newsong":  "1",
	}

	res, err := self.request("GET", "https://c.y.qq.com/v8/fcg-bin/fcg_v8_album_detail_cp.fcg", data)
	if err != nil {
		return nil, err
	}

	result := gjson.ParseBytes(res)

	album := &music.Album{
		Id:   result.Get("data.getAlbumInfo.Falbum_mid").String(),
		Name: result.Get("data.getAlbumInfo.Falbum_name").String(),
	}
	if sgs := result.Get("data.getSongInfo").Array(); len(sgs) > 0 {
		songs := make([]*music.Song, len(sgs))
		for i, sg := range sgs {
			songs[i] = self.toSong(sg)
		}
		album.Songs = songs
	}
	return album, nil
}

func (self *tencent) Lyric(id string) (*music.Lyric, error) {
	data := map[string]string{
		"songmid":  id,
		"g_tk":     "5381",
		"format":   "json",
		"platform": "yqq",
	}

	res, err := self.request("GET", "https://c.y.qq.com/lyric/fcgi-bin/fcg_query_lyric_new.fcg", data)
	if err != nil {
		return nil, err
	}

	result := gjson.ParseBytes(res)

	lrc := &music.Lyric{
		Lyric: base64Decode(result.Get("lyric").String()),
		Trans: base64Decode(result.Get("trans").String()),
	}
	return lrc, nil
}

func (self *tencent) Artist(id string) (*music.Artist, error) {
	data := map[string]string{
		"singermid": id,
		"begin":     "0",
		"num":       "50",
		"order":     "listen",
		"platform":  "yqqq",
		"newsong":   "1",
	}

	res, err := self.request("GET", "https://c.y.qq.com/v8/fcg-bin/fcg_v8_singer_track_cp.fcg", data)
	if err != nil {
		return nil, err
	}

	result := gjson.ParseBytes(res)

	ins := &music.Artist{
		Id:   result.Get("data.singer_mid").String(),
		Name: result.Get("data.singer_name").String(),
	}
	if sgs := result.Get("data.list").Array(); len(sgs) > 0 {
		songs := make([]*music.Song, len(sgs))
		for i, sg := range sgs {
			songs[i] = self.toSong(sg)
		}
		ins.Songs = songs
	}
	return ins, nil
}

func (self *tencent) Playlist(id string) (*music.Playlist, error) {
	data := map[string]string{
		"id":       id,
		"format":   "json",
		"newsong":  "1",
		"platform": "jqspaframe.json",
	}

	res, err := self.request("GET", "https://c.y.qq.com/v8/fcg-bin/fcg_v8_playlist_cp.fcg", data)
	if err != nil {
		return nil, err
	}

	result := gjson.ParseBytes(res)

	ins := &music.Playlist{
		Id:   result.Get("data.cdlist.0.disstid").String(),
		Name: result.Get("data.cdlist.0.dissname").String(),
	}
	if sgs := result.Get("data.cdlist.0.songlist").Array(); len(sgs) > 0 {
		songs := make([]*music.Song, len(sgs))
		for i, sg := range sgs {
			songs[i] = self.toSong(sg)
		}
		ins.Songs = songs
	}
	return ins, nil
}

func (self *tencent) Search(keyword string) ([]*music.Song, error) {
	data := map[string]string{
		"w":        keyword,
		"p":        "1",
		"n":        "30",
		"aggr":     "1",
		"lossless": "1",
		"cr":       "1",
		"new_json": "1",
		"format":   "json",
	}

	res, err := self.request("GET", "https://c.y.qq.com/soso/fcgi-bin/client_search_cp", data)
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
	music.Register("tencent", func(config *music.Config, client *http.Client) music.Server {
		return &tencent{
			config: config,
			client: client,
		}
	})
}
