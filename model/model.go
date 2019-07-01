package model

const (
	DEFAULT_R     float64 = 1500
	DEFAULT_RD    float64 = 350
	DEFAULT_SIGMA float64 = 0.06
)

type Page struct {
	Title string
	Body  []byte
}

type Pool struct {
	PoolName string   `json:"poolName"`
	Players  []Player `json:"players"`
}

type Player struct {
	Name     string `json:"name"`
	Glicko   Glicko `json:"glicko"`
	Absences int    `json:"absences"`
}

type Glicko struct {
	R   float64 `json:"r"`
	Rd  float64 `json:"rd"`
	Sig float64 `json:"sig"`
}

type EditPlayersReq struct {
	Name           string   `json:"name"`
	NewPlayers     []string `json:"newP"`
	RemovedPlayers []string `json:"remP"`
}

type GlickoEditReq struct {
	ChallongeURL string `json:"apiURL"`
	Name         string `json:"name"`
}

// GET request to Challonge
type ChallongeReq struct {
	Participants []ChallongeParticipant
	Matches      []ChallongeMatch
	IDMap        map[int]string
}

type ChallongeParticipant struct {
	Participant P `json:"participant"`
}
type P struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Seed      int    `json:"seed"`
	Placement int    `json:"final_rank"`
}

type ChallongeMatch struct {
	Match M `json:"match"`
}
type M struct {
	P1       int `json:"player1_id"`
	P2       int `json:"player2_id"`
	WinnerID int `json:"winner_id"`
}

type TournamentReq struct {
	Players []string `json:"players"`
	Name    string   `json:"name"`
}

/* Used for Glicko calculations.
 * Couples a Player with their Challonge ID.
 */
type GlickoPlayer struct {
	player *Player
	ID     int
}

func NewPlayer(name string) *Player {
	p := Player{name, Glicko{DEFAULT_R, DEFAULT_RD, DEFAULT_SIGMA}, 0}
	return &p
}

func NewGlickoPlayer(p *Player, id int) *GlickoPlayer {
	gp := GlickoPlayer{p, id}
	return &gp
}

func (g GlickoPlayer) Player() *Player {
	return g.player
}

func (p *Player) SetPlayerGlicko(r, rd, sig float64) {
	p.Glicko.R = r
	p.Glicko.Rd = rd
	p.Glicko.Sig = sig
}

func (p *Player) IncreaseAbsence() {
	p.Absences += 1
}

func (p *Player) ResetAbsences() {
	p.Absences = 0
}
