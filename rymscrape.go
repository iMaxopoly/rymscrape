package main

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

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

		var underSelection *goquery.Selection
		if rym.jseed.FullListAcquire.Under != "" {
			debugLog("Scoping under", rym.jseed.FullListAcquire.Under)
			underSelection = doc.Find(rym.jseed.FullListAcquire.Under)
		} else {
			underSelection = doc.Selection
		}

		if len(rym.jseed.FullListAcquire.LookFor) <= 0 {
			debugLog("rym.jseed.FullListAcquire.LookFor <= 0")
			continue
		}

		var lookForSelection *goquery.Selection
		for i, lookFor := range rym.jseed.FullListAcquire.LookFor {
			switch i {
			case 0:
				lookForSelection = underSelection.Find(lookFor)
			case len(rym.jseed.FullListAcquire.LookFor) - 1:
				lookForSelection = lookForSelection.Find(lookFor)
				for node := range lookForSelection.Nodes {
					res, exists := lookForSelection.Eq(node).Attr(rym.jseed.FullListAcquire.Res)
					if !exists || strings.TrimSpace(res) == "" {
						debugLog("rym.jseed.FullListAcquire.Res empty")
						continue
					}
					infoLog("Res", res)
				}
			default:
				lookForSelection = lookForSelection.Find(lookFor)
			}
		}
	}
}
