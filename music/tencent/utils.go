package tencent

import (
	"encoding/base64"
	"io"
	"math/rand"
	"net/url"
	"strings"
	"time"
)

func formData(data map[string]string) io.Reader {
	form := url.Values{}
	for k, v := range data {
		form.Set(k, v)
	}
	return strings.NewReader(form.Encode())
}

func base64Decode(s string) string {
	result, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return ""
	}
	return string(result)
}

func randomBytes(length int, charset string) []byte {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return b
}
