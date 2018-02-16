package main

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/blendlabs/go-logger"
)

type message struct {
	Time   time.Time `json:"time"`
	Status int       `json:"status,omitempty"`
	URL    string    `json:"url"`
	Error  string    `json:"error,omitempty"`
}

func main() {
	log := logger.All().WithWriter(logger.NewJSONWriterFromEnv())
	urls := strings.Split(os.Getenv("URLS"), ",")
	for _, url := range urls {
		go func(url string) {
			for {
				resp, err := http.Get(url)
				if err != nil {
					log.Write(wrap(url, 0, err.Error()))
				}
				log.Write(wrap(url, resp.StatusCode, ""))
				time.Sleep(time.Millisecond * 500)
			}
		}(url)
	}
	select {}
}

func wrap(url string, status int, err string) message {
	return message{
		Time:   time.Now().UTC(),
		Status: status,
		URL:    url,
		Error:  err,
	}
}

func (m message) Flag() logger.Flag {
	return "ping"
}

func (m message) Timestamp() time.Time {
	return m.Time
}
