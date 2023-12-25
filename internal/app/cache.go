package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/honmaple/forest"
	"github.com/nutsdb/nutsdb"
)

const bucket = "_default"

type Response struct {
	Code   int
	Body   []byte
	Header http.Header
}

type responseWriter struct {
	http.ResponseWriter
	multiWriter io.Writer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	return w.multiWriter.Write(b)
}

func (app *App) initCache() error {
	path := app.Config.GetString("cache.path")
	if path == "" {
		return nil
	}

	opt := nutsdb.DefaultOptions
	opt.SegmentSize = 8 * nutsdb.MB
	opt.CommitBufferSize = 4 * nutsdb.MB
	opt.MaxBatchSize = (15 * opt.SegmentSize / 4) / 100
	opt.MaxBatchCount = (15 * opt.SegmentSize / 4) / 100 / 100
	cache, err := nutsdb.Open(opt, nutsdb.WithDir(path))
	if err != nil {
		return err
	}
	cache.Update(func(tx *nutsdb.Tx) error {
		return tx.NewBucket(nutsdb.DataStructureBTree, bucket)
	})
	app.Cache = cache
	return nil
}

func (app *App) cacheResponse(c forest.Context) error {
	req := c.Request()
	if app.Cache == nil || req.Method != "GET" {
		return c.Next()
	}

	var (
		key   = fmt.Sprintf("%s:%s", req.Method, req.URL.String())
		value []byte
	)
	err := app.Cache.View(func(tx *nutsdb.Tx) error {
		v, err := tx.Get(bucket, []byte(key))
		if err != nil {
			return err
		}
		value = v
		return nil
	})
	if err == nil && value != nil {
		var resp Response

		if err := json.Unmarshal(value, &resp); err != nil {
			app.Log.Error(err.Error())
		} else {
			r := c.Response()
			for k, vs := range resp.Header {
				for _, v := range vs {
					r.Header().Add(k, v)
				}
			}
			if len(resp.Body) > 0 {
				r.Write(resp.Body)
			}
			r.WriteHeader(resp.Code)
			return nil
		}
	}

	var buf bytes.Buffer

	w := c.Response().ResponseWriter
	c.Response().ResponseWriter = &responseWriter{ResponseWriter: w, multiWriter: io.MultiWriter(w, &buf)}

	defer func() {
		resp := Response{
			Body:   buf.Bytes(),
			Code:   c.Response().Status,
			Header: c.Response().Header().Clone(),
		}
		b, err := json.Marshal(&resp)
		if err != nil {
			app.Log.Error(err.Error())
			return
		}
		var ttl uint32 = app.Config.GetUint32("cache.ttl")
		if resp.Code >= 400 {
			ttl = 60
		}
		err = app.Cache.Update(func(tx *nutsdb.Tx) error {
			return tx.Put(bucket, []byte(key), b, ttl)
		})
		if err != nil {
			app.Log.Error(err.Error())
		}
	}()
	return c.Next()
}
