package main

import (
	"github.com/ec16431/autournament/model"
	"github.com/zelenin/go-glicko2"
	"strings"
)

func calculateGlicko(req model.ChallongeReq, pool *model.Pool) []model.Player {
	participants := generatePlayerList(req, pool) // Get list of players to calculate
	idToGlicko := make(map[int]*glicko.Player)    // Maps ID -> Players

	for _, p := range participants {
		pGlicko := p.Player().Glicko
		idToGlicko[p.ID] = glicko.NewPlayer(glicko.NewRating(pGlicko.R, pGlicko.Rd, pGlicko.Sig))
	}

	period := glicko.NewRatingPeriod()

	for _, m := range req.Matches {

		var mRes glicko.MatchResult
		if m.Match.WinnerID == m.Match.P1 {
			mRes = glicko.MATCH_RESULT_WIN
		} else {
			mRes = glicko.MATCH_RESULT_LOSS
		}

		period.AddMatch(idToGlicko[m.Match.P1], idToGlicko[m.Match.P2], mRes)
	}

	period.Calculate()

	var retPlayers []model.Player

	// Update Player's Glicko ratings
	for _, p := range participants {
		rating := idToGlicko[p.ID].Rating()
		p.Player().SetPlayerGlicko(rating.R(), rating.Rd(), rating.Sigma())

		retPlayers = append(retPlayers, *p.Player())
	}

	return retPlayers
}

// Return list of players, both new and existing ones from the Challonge req
func generatePlayerList(r model.ChallongeReq, pool *model.Pool) []*model.GlickoPlayer {

	poolPlayers := loadPlayerMap(pool.Players) // Map of all players in existing pool
	var retPlayers []*model.GlickoPlayer       // All players (new and existing) with glicko raitngs

	// Load retPlayers with either new players or existing ones from Challonge req
	for _, p := range r.Participants {
		nu := strings.ToUpper(p.Participant.Name)
		val, exists := poolPlayers[nu]

		if exists {
			retPlayers = append(retPlayers, model.NewGlickoPlayer(&val.player, p.Participant.ID)) // Existing player
		} else {
			retPlayers = append(retPlayers, model.NewGlickoPlayer(model.NewPlayer(p.Participant.Name), p.Participant.ID)) // New player
		}
	}

	return retPlayers
}

// Return map of [playerName] -> Player object
func loadPlayerMap(players []model.Player) map[string]playerMapEntry {
	r := make(map[string]playerMapEntry)

	for _, p := range players {
		tu := strings.ToUpper(p.Name)
		r[tu] = playerMapEntry{tu, p}
	}
	return r
}

type playerMapEntry struct {
	nameUpper string
	player    model.Player
}
