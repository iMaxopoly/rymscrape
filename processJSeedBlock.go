package main

import (
	"errors"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// processSeedBlock processes the seed block and gets the target attribute content
// for example in the following seed snippet:
//
//      episodeListAcquire
//          under = #videos
//          lookFor[] = li, a
//          res = href
//
// The following functions finds #videos and then goes through all li elements. Then the first child element a is sought
// and finally the href attribute is stored.
func processSeedBlock(doc *goquery.Document, lookFor []string, under, res, siteProtocol, siteLink string) ([]string, error) {
	var underSelection *goquery.Selection
	if under != "" {
		debugLog("Scoping under", under)
		underSelection = doc.Find(under)
	} else {
		underSelection = doc.Selection
	}

	if len(lookFor) <= 0 {
		debugLog("LookFor <= 0")
		return []string{}, errors.New("lookFor empty")
	}

	var resCollection []string

	var lookForSelection *goquery.Selection
	for i, lf := range lookFor {
		switch i {
		case 0:
			lookForSelection = underSelection.Find(lf)

		case len(lookFor) - 1:
			lookForSelection = lookForSelection.Find(lf)
			for node := range lookForSelection.Nodes {
				res, exists := lookForSelection.Eq(node).Attr(res)
				if !exists || strings.TrimSpace(res) == "" {
					debugLog("res empty")
					continue
				}

				res = strings.TrimSpace(res)
				if !strings.HasPrefix(res, "htt") {
					res = siteProtocol + "://" + siteLink + "/" + res
				}

				infoLog("Acquiring", res)
				resCollection = append(resCollection, res)
			}

		default:
			lookForSelection = lookForSelection.Find(lf)
		}
	}

	return resCollection, nil
}
