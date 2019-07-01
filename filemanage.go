package main

import (
	"encoding/json"
	"fmt"
	"github.com/ec16431/autournament/model"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

// List all pools in the /pools/ directory
func listPools() []string {
	var pools []string
	r := regexp.MustCompile(".json")

	dir, _ := ioutil.ReadDir("./pools")
	for _, f := range dir {
		if f.Name() == "playerSchema.json" {
			continue
		}

		pools = append(pools, r.ReplaceAllString(f.Name(), ""))
	}
	return pools
}

// Reads JSON pool and returns JSON object for it
func readPool(name string) (*model.Pool, error) {
	// Read JSON
	jsonFile, err := os.Open("pools/" + name + ".json")
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()

	var res model.Pool
	err = json.NewDecoder(jsonFile).Decode(&res)
	if err != nil {
		log.Fatal(err)
	}

	return &res, nil
}

// Save a list of players to a pool and return the pool's name.
func saveJson(jsonDat []byte, name string) (string, error) {
	//Save to file
	fName := strings.Replace(name, " ", "-", -1) // Remove spaces
	pName := strings.ToLower(fName)
	fName = fmt.Sprintf("pools/%s.json", pName)
	err := ioutil.WriteFile(fName, jsonDat, 0644)
	if err != nil {
		return "", err
	}
	return pName, nil
}

func writeTestResults(n string, d []byte) error {
	fn := fmt.Sprintf("testRes/%s.txt", n)
	return ioutil.WriteFile(fn, d, 0644)
}
