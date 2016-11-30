package main

import "github.com/PuerkitoBio/goquery"

// rymscrape is the core of the project where:
// workers = number of concurrent goroutines.
// timeout = per request based timeout for http connections.
// timestamp = the timestamp at which procedure was started.
type rymscrape struct {
	workers   uint
	timeout   uint
	timestamp string
	jseed     JSeed
}

// getFullList parses through the jseed file and operates based on the commands given to fetch
// the target full brand page links.
func (rym *rymscrape) getFullList() (fullListLinks []string) {
	if len(rym.jseed.FullListLinks) <= 0 {
		debugLog("rym.jseed.FullListLinks <= 0")
		return
	}

	for _, fullLink := range rym.jseed.FullListLinks {
		debugLog("Discovered fullLink entity", fullLink)
		fullLink = rym.jseed.SiteProtocol + "://" + rym.jseed.SiteLink + "/" + fullLink

		doc, err := goquery.NewDocument(fullLink)
		if err != nil {
			errorLog(err)
			continue
		}

		links, err := processSeedBlock(
			doc,
			rym.jseed.FullListAcquire.LookFor,
			rym.jseed.FullListAcquire.Under,
			rym.jseed.FullListAcquire.Res,
			rym.jseed.SiteProtocol,
			rym.jseed.SiteLink,
		)
		if err != nil {
			errorLog(err)
			continue
		}

		fullListLinks = append(fullListLinks, links...)
	}

	return fullListLinks
}

// getEpisodeList parses through the jseed file and operates based on the commands given to fetch
// the target full brand episode links from the brand page links provided.
func (rym *rymscrape) getEpisodeList(brandLink string) (episodeLinks []string, err error) {
	doc, err := goquery.NewDocument(brandLink)
	if err != nil {
		errorLog(err)
		return []string{}, err
	}

	episodeLinks, err = processSeedBlock(
		doc,
		rym.jseed.EpisodeListAcquire.LookFor,
		rym.jseed.EpisodeListAcquire.Under,
		rym.jseed.EpisodeListAcquire.Res,
		rym.jseed.SiteProtocol,
		rym.jseed.SiteLink,
	)

	return episodeLinks, nil
}

// getVideoList parses through the jseed file and operates based on the commands given to fetch
// the video links from the episode link provided.
func (rym *rymscrape) getVideoList(episodeLink string) (videoLinks []string, err error) {

	return videoLinks, nil
}
