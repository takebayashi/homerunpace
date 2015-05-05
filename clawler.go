package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/takebayashi/npbbis"
)

func crawl(date string) {
	games, _ := npbbis.GetGames(date)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	for _, game := range games {
		gExists, err := db.Query("SELECT 1 FROM games WHERE id = $1", game.Id)
		if err != nil {
			panic(err)
		}
		gInsert := "INSERT INTO games VALUES ($1, $2, $3, 0)"
		if gExists.Next() {
			gInsert = "UPDATE games SET date = $2, status = $3 WHERE id = $1"
		}
		_, err = db.Exec(gInsert, game.Id, game.Date, game.Status)
		if err != nil {
			panic(err)
		}
		for _, hr := range game.Homeruns {
			hrExists, err := db.Query("SELECT 1 FROM homeruns WHERE game = $1 AND batter = $2 AND number = $3", game.Id, hr.Batter, hr.Number)
			if err != nil {
				panic(err)
			}
			hrInsert := "INSERT INTO homeruns VALUES ($1, $2, $3, $4, $5)"
			if hrExists.Next() {
				hrInsert = "UPDATE homeruns SET scenario = $4, pitcher = $5 WHERE game = $1 AND batter = $2 AND number = $3"
			}
			_, err = db.Exec(hrInsert, game.Id, hr.Batter, hr.Number, hr.Scenario, hr.Pitcher)
			if err != nil {
				panic(err)
			}
		}
	}
}
