package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/takebayashi/npbbis"
)

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
