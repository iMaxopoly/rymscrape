package main

import "github.com/PuerkitoBio/goquery"

type rymscrape struct {
	workers   uint
	timeout   uint
	timestamp string
	jseed     JSeed
}

func (rym *rymscrape) getFullList() {
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

		if rym.jseed.FullListAcquire.Under != "" {
			debugLog("Scoping under", rym.jseed.FullListAcquire.Under)
			doc = doc.Find(rym.jseed.FullListAcquire.Under)
		}

		if len(rym.jseed.FullListAcquire.LookFor) <= 0 {
			debugLog("rym.jseed.FullListAcquire.LookFor <= 0")
			continue
		}

		for _, lookFor := range rym.jseed.FullListAcquire.LookFor {
			doc.Find(lookFor)
		}
	}
}
