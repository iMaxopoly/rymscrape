package main

import (
	"errors"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/uber-go/zap"
)

// processSeedBlock processes the seed block and gets the target attribute content
// for example in the following seed snippet:
//
//      episodeListAcquire
//          isTrue b = true
//          paginate
//              isTrue b = false
//              under = nil
//              lookFor[] = nil
//              res = nil
//          under = #videos
//          lookFor[] = li, a
//          res = href
//
// The following functions finds #videos and then goes through all li elements. Then the first child element a is sought
// and finally the href attribute is stored.
func processSeedBlock(doc *goquery.Document, lookFor []string, under, res, siteProtocol, siteLink string) ([]string, error) {
	var underSelection *goquery.Selection
	if under != "" {
		logger.Debug("Scoping under", zap.String("under", under))
		underSelection = doc.Find(under)
	} else {
		underSelection = doc.Selection
	}

	if len(lookFor) <= 0 {
		logger.Debug("LookFor <= 0")
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
					logger.Debug("res empty")
					continue
				}

				res = strings.TrimSpace(res)
				if !strings.HasPrefix(res, "//") && !strings.HasPrefix(res, "htt") {
					if strings.HasPrefix(res, "/") {
						res = siteProtocol + "://" + siteLink + res
					} else {
						res = siteProtocol + "://" + siteLink + "/" + res
					}
				}

				if strings.HasPrefix(res, "//www.dailymotion.com") {
					res = strings.Replace(res, "//www.dailymotion.com", "http://www.dailymotion.com", 1)
				} else if strings.HasPrefix(res, "//www.youtube.com") {
					res = strings.Replace(res, "//www.youtube.com", "https://www.youtube.com", 1)
				}

				logger.Info("Acquiring", zap.String("res", res))
				resCollection = append(resCollection, res)
			}

		default:
			lookForSelection = lookForSelection.Find(lf)
		}
	}

	return resCollection, nil
}
