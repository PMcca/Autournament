package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strings"
)

var templates *template.Template

func httpInit() {
	// Handler init
	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/new-pool/", newPoolHandler)
	http.HandleFunc("/pool/", poolHandler)
	http.HandleFunc("/save/", saveHandler)
	http.HandleFunc("/pool/edit/", editHandler)
	http.HandleFunc("/pool/glicko/", glickoHandler)
	http.HandleFunc("/pool/tournament/", tournamentHandler)

	// Fileserver init
	http.Handle("/pools/", http.StripPrefix("/pools/", http.FileServer(http.Dir("pools"))))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Template init
	templates = template.Must(template.ParseFiles("tmpl/main.html", "tmpl/new-pool.html", "tmpl/pool.html"))
}

// Call API and train dataset
func trainSet(poolName string) {
	// MAKE SURE EMPTY POOL EXISTS

	var urls []string
	suffix := "?api_key=hRnTyyGav36S6bxcJzhY37kQktPrHs33mmTUXjti&include_participants=1&include_matches=1"

	// Non-prefixed ones
	for i := 3; i < 10; i++ {
		for j := 0; j < 10; j++ {

			// Special cases
			if i == 3 && j == 0 {
				continue
			}
			if i == 3 && j == 7 {
				continue
			}

			url := fmt.Sprintf("https://api.challonge.com/v1/tournaments/4qm%d%d%s", i, j, suffix)
			urls = append(urls, url)
		}
	}

	//100 - 106
	for i := 0; i < 5; i++ {
		url := fmt.Sprintf("https://api.challonge.com/v1/tournaments/4qm10%d%s", i, suffix)
		urls = append(urls, url)
	}

	// Prefixed ones
	for i := 0; i < 3; i++ {
		for j := 0; j < 10; j++ {

			if i == 0 && j < 5 {
				continue
			}

			if i == 2 && j == 2 {
				break
			}

			url := fmt.Sprintf("https://api.challonge.com/v1/tournaments/4qs-4qm1%d%d%s", i, j, suffix)
			urls = append(urls, url)
		}
	}

	for _, url := range urls {
		println(url)
	}

	// Calculate glicko for each tournament URL
	for _, url := range urls {
		pool, err := readPool(poolName)
		if err != nil {
			log.Fatal(err)
		}
		println("CALCULATING ", url)

		glicko := calculateGlicko(getTournament(url), pool)

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

		_, err = saveJson(j, poolName)
		if err != nil {
			log.Fatal(err)
		}

		println("SUCCESS FOR ", url)
	}

}

// Writes the seedings and placements for a given tournament in URL.
func writeSeedsAndPlacements(url, n string) {
	r := getTournament(url)
	var d strings.Builder
	d.WriteString("Name  Seed  Placement\n")

	for _, s := range r.Participants {
		d.WriteString(fmt.Sprintf("%s  %d  %d\n", s.Participant.Name, s.Participant.Seed, s.Participant.Placement))
	}

	writeTestResults(n, []byte(d.String()))
}

func main() {
	//trainSet("4-quarters")
	//writeSeedsAndPlacements("https://api.challonge.com/v1/tournaments/4qs-4qm132?api_key=hRnTyyGav36S6bxcJzhY37kQktPrHs33mmTUXjti&include_participants=1&include_matches=1", "132")

	log.SetFlags(log.Llongfile)
	httpInit()
	println("Listening on 8080...")
	http.ListenAndServe(":8080", nil)
}
