package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"strings"
	"unisearcher/functions"
)

//UniInfoHandler
func UniInfoHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == http.MethodGet {
		handleGetRequest(w, r)
	} else {
		http.Error(w, "Method not supported. Currently only GET are supported.", http.StatusNotImplemented)
		return
	}
}

/*
Dedicated handler for GET requests
*/
func handleGetRequest(w http.ResponseWriter, r *http.Request) {

	// Space friendly query :)
	query := strings.Replace(path.Base(r.URL.Path), " ", "%20", -1)

	// URL to invoke
	url := fmt.Sprintf("http://universities.hipolabs.com/search?name=%s", query)

	uniRequest := sendRequest(url)

	unis := make([]UniCache, 0)
	if err := uniRequest.Decode(&unis); err != nil {
		log.Fatal(err)
	}

	// Creates empty slice
	cca2 := make([]string, 0)
	// Adds isocodes into a slice, ignores duplicates
	for _, uni := range unis {
		if !functions.Contains(cca2, uni.AlphaTwoCode) {
			cca2 = append(cca2, uni.AlphaTwoCode)
		}
	}

	url = fmt.Sprintf("https://restcountries.com/v3.1/alpha?codes=%s", strings.Join(cca2[:], ","))

	if len(unis) != 0 {
		//Issues new request
		countryRequest := sendRequest(url)

		//Creates empty slice
		countryCache := make([]CountryCache, 0)

		//Decodes request
		if err := countryRequest.Decode(&countryCache); err != nil {
			log.Fatal(err)
		}

		//Creates an empy slice wit
		out := make([]UniInfoResponse, 0, len(unis))

		//Uses information from UniCache and CountryCache to create a new struct with the correct fields
		for _, obj := range unis {
			for _, c := range countryCache {
				if c.CCA2 == obj.AlphaTwoCode {
					out = append(out, UniInfoResponse{
						Name:      obj.Name,
						Country:   obj.Country,
						IsoCode:   obj.AlphaTwoCode,
						WebPages:  obj.WebPages,
						Languages: c.Languages,
						Map:       c.Maps["openStreetMaps"],
					})
				}
			}
		}

		sendResponse(w, out)
	} else {
		http.Error(w, "Results not found", http.StatusNoContent)
	}
}

func sendResponse(w http.ResponseWriter, r []UniInfoResponse) {
	// Write content type header
	w.Header().Add("content-type", "application/json")

	// Instantiate encoder
	encoder := json.NewEncoder(w)

	//Encodes response
	err := encoder.Encode(r)
	if err != nil {
		http.Error(w, "Error during encoding", http.StatusInternalServerError)
		return
	}
}

func sendRequest(url string) *json.Decoder {
	// Create new request
	r, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		_ = fmt.Errorf("error in creating request", err.Error())
	}

	// Instantiate the client
	client := &http.Client{}

	// Setting content type -> effect depends on the service provider
	r.Header.Add("content-type", "application/json")

	// Issue request
	res, err := client.Do(r)
	if err != nil {
		_ = fmt.Errorf("Error in response", err.Error())
	}

	decoder := json.NewDecoder(res.Body)

	return decoder
}
