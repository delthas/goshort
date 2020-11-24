//go:generate picopacker index.html index.go index
//go:generate picopacker api.html api.go api
package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

func randomHex(n int) string {
	bytes := make([]byte, n+1/2)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)[:n]
}

func add(w http.ResponseWriter, key string, url string, api bool) {
	if len(key) == 0 || len(url) == 0 {
		return
	}

	i := strings.Index(url, ":/")
	if i >= 0 && len(url) > i+2 && url[i+2] != '/' {
		url = url[:i] + "://" + url[i+2:]
	}

	if key == "hash" {
		key = randomHex(8)
	}

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("links"))
		err := b.Put([]byte(key), []byte(url))
		return err
	})
	if err != nil {
		w.WriteHeader(500)
		return
	}

	var t *template.Template
	if api {
		t = apiTemplate
	} else {
		t = indexTemplate
	}
	t.Execute(w, struct {
		Key string `html:"key"`
		URL string `html:"url"`
	}{
		Key: urlPrefix + key,
		URL: url,
	})
}

var indexTemplate = template.Must(template.New("index").Parse(string(index)))
var apiTemplate = template.Must(template.New("api").Parse(string(api)))

var port int
var urlPrefix string

var db *bolt.DB

func main() {
	rand.Seed(time.Now().UnixNano())

	flag.IntVar(&port, "port", 8080, "http server listen port")
	flag.StringVar(&urlPrefix, "url", "http://localhost:8080/", "http server external url prefix")
	flag.Parse()

	var err error
	db, err = bolt.Open("goshort.db", 0600, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucket([]byte("links"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if r.Method == "POST" {
			key := r.FormValue("key")
			url := r.FormValue("url")
			if key == "" {
				key = "hash"
			}
			add(w, key, url, false)
			return
		}
		if r.Method != "GET" {
			w.WriteHeader(400)
			return
		}
		path := r.URL.Path
		if len(path) == 0 {
			w.WriteHeader(400)
			return
		}
		if path[0] == '/' {
			path = path[1:]
		}
		if path == "" {
			indexTemplate.Execute(w, nil)
			return
		}

		i := strings.IndexRune(path, '/')
		if i >= 0 && i < len(path)-1 {
			key := path[:i]
			url := path[i+1:]
			add(w, key, url, true)
		} else {
			key := path
			if i > 0 {
				key = path[:i]
			}
			if key == "" {
				w.WriteHeader(404)
				return
			}

			var buf []byte
			db.View(func(tx *bolt.Tx) error {
				b := tx.Bucket([]byte("links"))
				v := b.Get([]byte(key))
				if v != nil {
					buf = make([]byte, len(v))
					copy(buf, v)
				}
				return nil
			})
			if buf == nil {
				w.WriteHeader(404)
				return
			}

			url := string(buf)
			if !strings.Contains(url, "://") {
				url = "http://" + url
			}
			w.Header().Set("Location", url)
			w.WriteHeader(302)
		}
	})
	fmt.Println("listening on " + urlPrefix)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}
