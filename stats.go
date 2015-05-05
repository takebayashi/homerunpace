package main

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type Stat struct {
	Date         string `json:"date"`
	GameCount    int    `json:"game_count"`
	HomerunCount int    `json:"homerun_count"`
}

func getStats(year int) ([]*Stat, error) {
	db, err := sql.Open("postgres", dsn)
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
	db, err := sql.Open("postgres", dsn)
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
