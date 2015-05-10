package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"github.com/Sirupsen/logrus"
	"github.com/garyburd/redigo/redis"
	"github.com/goji/glogrus"
	_ "github.com/lib/pq"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
	"net/http"
	"os"
	"strconv"
	"time"
)

var dsn string
var updateToken string

var redisNetwork string
var redisAddress string
var redisPassword string
var redisMaxIdle int

var db *sql.DB

func main() {
	var err error

	// configure
	logFile := os.Getenv("HOMERUNPACE_LOGFILE")
	dsn = os.Getenv("HOMERUNPACE_DSN")
	updateToken = os.Getenv("HOMERUNPACE_TOKEN")
	redisNetwork = os.Getenv("HOMERUNPACE_REDIS_NETWORK")
	redisAddress = os.Getenv("HOMERUNPACE_REDIS_ADDRESS")
	redisPassword = os.Getenv("HOMERUNPACE_REDIS_PASSWORD")
	var envErr error
	redisMaxIdle, envErr = strconv.Atoi("HOMERUNPACE_REDIS_MAX_IDLE")
	if envErr != nil {
		redisMaxIdle = 5
	}
	root := os.Getenv("HOMERUNPACE_ROOT")

	// connect to database
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// build server
	flag.Set("bind", ":80")
	goji.Get("/stats/:year", handleStats)
	goji.Post("/crawler/:date", handleCrawling)
	goji.Handle("/", http.FileServer(http.Dir(root+"/static")))

	// replace default logger with Logrus
	goji.Abandon(middleware.Logger)
	logger := logrus.New()
	logger.Formatter = new(logrus.JSONFormatter)
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		logger.Out = f
	}
	goji.Use(glogrus.NewGlogrus(logger, "homerunpace"))

	// start app
	goji.Serve()
}

var redisPool *redis.Pool

func newRedisConn() redis.Conn {
	if redisPool == nil {
		redisPool = &redis.Pool{
			MaxIdle:     redisMaxIdle,
			IdleTimeout: 600 * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial(redisNetwork, redisAddress)
				if err != nil {
					return nil, err
				}
				if redisPassword != "" {
					_, err = c.Do("AUTH", redisPassword)
					if err != nil {
						c.Close()
						return nil, err
					}
				}
				return c, err
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
		}
	}
	return (*redisPool).Get()
}

func handleStats(c web.C, w http.ResponseWriter, r *http.Request) {
	year, _ := strconv.Atoi(c.URLParams["year"])
	by := r.FormValue("by")
	key := strconv.Itoa(year) + by
	var j []byte

	rc := newRedisConn()
	defer rc.Close()

	v, err := redis.String(rc.Do("GET", key))
	if err != nil {
		j = generateStats(year, by)
		rc.Do("SET", key, j)
	} else {
		j = []byte(v)
	}
	w.Write(j)
}

func generateStats(year int, by string) []byte {
	var stats []interface{}
	switch by {
	case "date":
		ss, _ := getStats(year)
		for _, s := range ss {
			stats = append(stats, s)
		}
	case "game":
		ss, _ := getGameBaseStats(year)
		for _, s := range ss {
			stats = append(stats, s)
		}
	}
	j, _ := json.MarshalIndent(stats, "", "\t")
	return j
}

func handleCrawling(c web.C, w http.ResponseWriter, r *http.Request) {
	if r.FormValue("token") == updateToken {
		date := c.URLParams["date"]
		crawl(date)
		year, _ := strconv.Atoi(date[:4])

		rc := newRedisConn()
		defer rc.Close()

		for _, by := range []string{"date", "game"} {
			key := strconv.Itoa(year) + by
			j := generateStats(year, by)
			rc.Do("SET", key, string(j))
		}
		w.Write([]byte("Update done: " + date + "\n"))
	} else {
		w.WriteHeader(403)
	}
}
