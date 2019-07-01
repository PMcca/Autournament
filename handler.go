package main

import (
	"encoding/json"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/ec16431/autournament/model"
)

func mainHandler(w http.ResponseWriter, r *http.Request) {
	page := model.Page{Title: "Autournament"}

	ctx := buildContext(page)
	renderTemplate(w, "main", ctx)
}

func newPoolHandler(w http.ResponseWriter, r *http.Request) {
	body := r.FormValue("body")
	page := model.Page{r.URL.Path, []byte(body)}

	ctx := buildContext(page)
	renderTemplate(w, "new-pool", ctx)
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	name := r.Form["name"]
	players := r.Form["player"]

	pool := model.Pool{PoolName: name[0]}
	for _, p := range players {
		if len(p) > 0 {
			player := *model.NewPlayer(p)
			pool.Players = append(pool.Players, player)
		}
	}

	poolJson, err := json.Marshal(pool)
	if err != nil {
		log.Fatal(err)
	}

	pName, err := saveJson(poolJson, pool.PoolName)
	if err != nil {
		log.Fatal(err)
	}

	page := model.Page{Title: name[0]}

	ctx := buildContext(page, pool)
	ctx["pName"] = pName
	renderTemplate(w, "pool", ctx)
}

func poolHandler(w http.ResponseWriter, r *http.Request) {
	pName := r.URL.Path[len("/pool/"):]
	if len(pName) == 0 {
		pools := listPools()
		page := model.Page{Title: "Pools"}

		ctx := buildContext(page)
		ctx["pools"] = pools
		ctx["listPools"] = true

		renderTemplate(w, "pool", ctx)
		return
	}
	pName = strings.ToLower(pName)

	res, err := readPool(pName)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	page := model.Page{Title: res.PoolName}
	ctx := make(map[string]interface{})
	ctx["pName"] = pName
	ctx["Pool"] = res
	ctx["Page"] = page

	renderTemplate(w, "pool", ctx)
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	var editReq model.EditPlayersReq
	err := json.NewDecoder(r.Body).Decode(&editReq)
	if err != nil {
		log.Fatal(err)
	}

	if len(editReq.NewPlayers) != 0 || len(editReq.RemovedPlayers) != 0 {

		pool, err := readPool(editReq.Name)
		if err != nil {
			log.Fatal(err)
		}

		// Remove players
		for _, r := range editReq.RemovedPlayers {
			for i, p := range pool.Players {
				if r == p.Name {
					pool.Players = append(pool.Players[:i], pool.Players[i+1:]...)
				}
			}
		}

		// Add players
		for _, p := range editReq.NewPlayers {
			if playerExists(p, *pool) {
				continue
			}
			pool.Players = append(pool.Players, *model.NewPlayer(p))
		}

		sort.Slice(pool.Players, func(i, j int) bool {
			return pool.Players[i].Glicko.R > pool.Players[j].Glicko.R
		})

		j, err := json.Marshal(pool)
		if err != nil {
			log.Fatal(err)
		}

		saveJson(j, editReq.Name)
	}

	w.WriteHeader(http.StatusOK)
}

func glickoHandler(w http.ResponseWriter, r *http.Request) {
	var glickoReq model.GlickoEditReq
	err := json.NewDecoder(r.Body).Decode(&glickoReq)
	if err != nil {
		log.Fatal(err)
	}

	pool, err := readPool(glickoReq.Name)
	if err != nil {
		log.Fatal(err)
	}

	glicko := calculateGlicko(getTournament(glickoReq.ChallongeURL), pool)

	// Merge new ranked players with ones that didn't participate
	addPlayer := true
	for _, p := range pool.Players {
		for j, gp := range glicko {
			if p.Name == gp.Name {
				addPlayer = false
				glicko[j].ResetAbsences() // Player attended
				break
			}
		}
		if addPlayer {
			p.IncreaseAbsence()
			glicko = append(glicko, p)
		}
		addPlayer = true
	}

	pool.Players = glicko
	sort.Slice(pool.Players, func(i, j int) bool {
		return pool.Players[i].Glicko.R > pool.Players[j].Glicko.R
	})

	j, err := json.Marshal(pool)
	if err != nil {
		log.Fatal(err)
	}

	_, err = saveJson(j, glickoReq.Name)
	if err != nil {
		log.Fatal(err)
	}

	w.WriteHeader(http.StatusOK)
}

func tournamentHandler(w http.ResponseWriter, r *http.Request) {
	var tournReq model.TournamentReq
	err := json.NewDecoder(r.Body).Decode(&tournReq)
	if err != nil {
		log.Fatal(err)
	}

	pool, err := readPool(tournReq.Name)
	if err != nil {
		log.Fatal(err)
	}

	var includedPlayers []model.Player

	for _, n := range tournReq.Players {
		for _, p := range pool.Players {
			if n == p.Name {
				includedPlayers = append(includedPlayers, p)
				break
			}
		}
	}

	// Sort based on how much their absences should influence their placement
	sort.Slice(includedPlayers, func(i, j int) bool {
		p1 := includedPlayers[i]
		p2 := includedPlayers[j]
		return p1.Glicko.R-float64((10*p1.Absences)) > p2.Glicko.R-float64((10*p2.Absences))
	})

	j, err := json.Marshal(includedPlayers)
	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}

func buildContext(objs ...interface{}) map[string]interface{} {
	r := regexp.MustCompile(`\w*\.`) // Remove dot-separated package prefixes

	var ctx = make(map[string]interface{})
	for _, o := range objs {
		n := r.ReplaceAllString(reflect.TypeOf(o).String(), (""))
		ctx[n] = o
	}
	return ctx
}

func renderTemplate(w http.ResponseWriter, tmpl string, context map[string]interface{}) {
	err := templates.ExecuteTemplate(w, tmpl+".html", context)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Return the list of participants and matches for a given challonge bracket
func getTournament(url string) model.ChallongeReq {
	tJson, err := challongeGetReq(url)
	if err != nil {
		log.Fatal(err)
	}
	pJson := gjson.Get(string(tJson), "tournament.participants")
	mJson := gjson.Get(string(tJson), "tournament.matches")

	// Participant list
	var participants []model.ChallongeParticipant
	err = json.Unmarshal([]byte(pJson.Raw), &participants)
	if err != nil {
		log.Fatal(err)
	}

	// Matches list
	var matches []model.ChallongeMatch
	err = json.Unmarshal([]byte(mJson.Raw), &matches)
	if err != nil {
		log.Fatal(err)
	}

	// Make ID to Name map
	// REFACTOR TO OWN FUNCTION
	IDMap := make(map[int]string)
	for _, p := range participants {
		IDMap[p.Participant.ID] = p.Participant.Name
	}

	res := model.ChallongeReq{participants, matches, IDMap}

	return res
}

// Return JSON dataset of a Challonge tournament bracket.
func challongeGetReq(url string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("ERROR- Constructing request: <%v>", err)
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Println("ERROR- Calling Challonge: ", err)
		return nil, err
	}

	defer resp.Body.Close()
	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("ERROR- Reading Challonge response: ", err)
		return nil, err
	}
	return contents, nil
}

// Returns true if player p exists in pool
func playerExists(p string, pool model.Pool) bool {
	for _, poolP := range pool.Players {
		if p == poolP.Name {
			return true
		}
	}

	return false
}
