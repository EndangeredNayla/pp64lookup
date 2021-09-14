package curseapi

import (
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/sync/singleflight"
)

var c = http.Client{Timeout: 10 * time.Second}

func httpget(url string) ([]byte, error) {
	reqs, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("httpget: %w", err)
	}
	reqs.Header.Set("Accept", "*/*")
	reqs.Header.Set("Accept-Encoding", "gzip")
	reqs.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/84.0.4147.105 Safari/537.36")
	rep, err := c.Do(reqs)
	if rep != nil {
		defer rep.Body.Close()
	}
	if err != nil {
		return nil, fmt.Errorf("httpget: %w", err)
	}
	var reader io.ReadCloser
	switch rep.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(rep.Body)
		if err != nil {
			return nil, fmt.Errorf("httpget: %w", err)
		}
		defer reader.Close()
	default:
		reader = rep.Body
	}
	b, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("httpget: %w", err)
	}
	return b, err
}

var acache = newcache(30 * time.Minute)

var s = singleflight.Group{}

func httpcache(url string) ([]byte, error) {
	b := acache.Load(url)
	if b != nil {
		return b, nil
	}
	t, err, _ := s.Do(url, func() (interface{}, error) {
		b, err := httpget(url)
		if err != nil {
			return nil, err
		}
		acache.Store(url, b)
		return b, nil
	})
	if err != nil {
		return nil, fmt.Errorf("httpcache: %w", err)
	}
	b = t.([]byte)
	return b, nil
}
