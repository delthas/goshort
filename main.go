//go:generate packer index.html index.go index
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

func main() {
	rand.Seed(time.Now().UnixNano())

	port := flag.Int("port", 8080, "http server listen port")
	urlPrefix := flag.String("url", "http://localhost:8080/", "http server external url prefix")
	flag.Parse()

	db, err := bolt.Open("goshort.db", 0600, nil)
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

	okTemplate := template.Must(template.New("ok").Parse(string(index)))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

		i := strings.IndexRune(path, '/')
		if i < 0 || i == len(path)-1 {
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
		} else {
			key := path[:i]
			url := path[i+1:]

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

			okTemplate.Execute(w, struct {
				Short string `html:"short"`
				Url   string `html:"url"`
			}{
				Short: *urlPrefix + key,
				Url:   url,
			})
		}

	})
	fmt.Println("listening on " + *urlPrefix)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*port), nil))
}
