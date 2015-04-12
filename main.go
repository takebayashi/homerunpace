package main

import (
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hisaichi5518/vache"
	"github.com/takebayashi/npbbis"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"net/http"
	"os"
	"strconv"
	"time"
)

var dsn string

func main() {
	dsn = os.Getenv("HOMERUNRATE_DSN")
	goji.Get("/stats/:year", handleStats)
	goji.Post("/crawler/:date", handleCrawling)
	goji.Handle("/", http.FileServer(http.Dir("static")))
	goji.Serve()
}

func handleStats(c web.C, w http.ResponseWriter, r *http.Request) {
	year, _ := strconv.Atoi(c.URLParams["year"])
	by := r.FormValue("by")
	key := strconv.Itoa(year) + by
	var j []byte
	v, ok := vache.Get(key)
	if !ok {
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
		j, _ = json.MarshalIndent(stats, "", "\t")
		vache.Set(key, string(j), 3600*time.Second)
	} else {
		j = []byte(v)
	}
	w.Header().Set("Content-Encoding", "gzip")
	gz := gzip.NewWriter(w)
	defer gz.Close()
	gz.Write(j)
}

type Stat struct {
	Date         string `json:"date"`
	GameCount    int    `json:"game_count"`
	HomerunCount int    `json:"homerun_count"`
}

func getStats(year int) ([]*Stat, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	q := `
    SELECT
      date,
      COUNT(*),
      (SELECT COUNT(*) FROM homeruns h INNER JOIN games g ON h.game = g.id WHERE YEAR(g.date) = ? AND g.date <= o.date)
    FROM games o
    WHERE YEAR(date) = ? AND status = "" AND type = 0 GROUP BY date ORDER BY date
  `
	rows, err := db.Query(q, year, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stats = []*Stat{}
	for rows.Next() {
		var date string
		var gCount int
		var hrCount int
		err := rows.Scan(&date, &gCount, &hrCount)
		if err != nil {
			return nil, err
		}
		stats = append(stats, &Stat{Date: date, GameCount: gCount, HomerunCount: hrCount})
	}
	return stats, nil
}

type GameBaseStat struct {
	GameCount    int    `json:"game_count"`
	GameId       string `json:"game_id"`
	HomerunCount int    `json:"homerun_count"`
}

func getGameBaseStats(year int) ([]*GameBaseStat, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	q := `
    SELECT id, (SELECT COUNT(*) FROM homeruns h INNER JOIN games g ON h.game = g.id WHERE YEAR(g.date) = ? AND g.id <= o.id) FROM games o WHERE YEAR(date) = ? AND status = "" AND type = 0 GROUP BY id ORDER BY id
  `
	rows, err := db.Query(q, year, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stats = []*GameBaseStat{}
	gCount := 0
	for rows.Next() {
		gCount++
		var gId string
		var hrCount int
		err := rows.Scan(&gId, &hrCount)
		if err != nil {
			return nil, err
		}
		stats = append(stats, &GameBaseStat{GameCount: gCount, GameId: gId, HomerunCount: hrCount})
	}
	return stats, nil
}

func handleCrawling(c web.C, w http.ResponseWriter, r *http.Request) {
	crawl(c.URLParams["date"])
	fmt.Fprintf(w, "done")
}

func crawl(date string) {
	games, _ := npbbis.GetGames(date)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	for _, game := range games {
		_, err := db.Exec("REPLACE INTO games VALUES (?, ?, ?, 0)", game.Id, game.Date, game.Status)
		if err != nil {
			panic(err)
		}
		for _, hr := range game.Homeruns {
			_, err := db.Exec("REPLACE INTO homeruns VALUES (?, ?, ?, ?, ?)", game.Id, hr.Batter, hr.Number, hr.Scenario, hr.Pitcher)
			if err != nil {
				panic(err)
			}
		}
	}
}
