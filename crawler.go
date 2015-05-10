package main

import (
	"github.com/takebayashi/npbbis"
)

func crawl(date string) {
	games, _ := npbbis.GetGames(date)
	for _, game := range games {
		updateGame(game, date)
		for _, hr := range game.Homeruns {
			updateHomerun(hr, game)
		}
	}
}

func updateGame(game *npbbis.Game, date string) error {
	exists, err := db.Query("SELECT 1 FROM games WHERE id = $1", game.Id)
	if err != nil {
		return err
	}
	defer exists.Close()
	q := "INSERT INTO games VALUES ($1, $2, $3, 0)"
	if exists.Next() {
		q = "UPDATE games SET date = $2, status = $3 WHERE id = $1"
	}
	_, err = db.Exec(q, game.Id, game.Date, game.Status)
	if err != nil {
		return err
	}
	return nil
}

func updateHomerun(hr *npbbis.Homerun, game *npbbis.Game) error {
	exists, err := db.Query("SELECT 1 FROM homeruns WHERE game = $1 AND batter = $2 AND number = $3", game.Id, hr.Batter, hr.Number)
	if err != nil {
		return err
	}
	defer exists.Close()
	q := "INSERT INTO homeruns VALUES ($1, $2, $3, $4, $5)"
	if exists.Next() {
		q = "UPDATE homeruns SET scenario = $4, pitcher = $5 WHERE game = $1 AND batter = $2 AND number = $3"
	}
	_, err = db.Exec(q, game.Id, hr.Batter, hr.Number, hr.Scenario, hr.Pitcher)
	if err != nil {
		return err
	}
	return nil
}
