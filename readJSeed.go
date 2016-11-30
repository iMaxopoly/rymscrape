package main

import (
	"encoding/json"
	"io/ioutil"
)

/*
JSeed is a struct that copies the json AES-512 encrypted file contents that describes the operations required to get links from.
*/
type JSeed struct {
	SiteLink        string   `json:"siteLink"`
	SiteProtocol    string   `json:"siteProtocol"`
	SiteSignature   string   `json:"siteSignature"`
	FullListLinks   []string `json:"fullListLinks"`
	FullListAcquire struct {
		Under   string   `json:"under"`
		LookFor []string `json:"lookFor"`
	} `json:"fullListAcquire"`
	EpisodeListAcquire struct {
		Under   string   `json:"under"`
		LookFor []string `json:"lookFor"`
	} `json:"episodeListAcquire"`
	VideoListAcquire struct {
		Under   string   `json:"under"`
		LookFor []string `json:"lookFor"`
	} `json:"videoListAcquire"`
}

func readJSeedFile() JSeed {
	raw, err := ioutil.ReadFile(*confSeedFile)
	handleErrorAndPanic(err)

	var j JSeed

	err = json.Unmarshal(raw, &j)
	handleErrorAndPanic(err)

	return j
}
