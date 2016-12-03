package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mozillazg/go-slugify"
	"github.com/uber-go/zap"
	"gopkg.in/go-playground/pool.v3"
)

// rymscrape is the core of the project where:
// workers = number of concurrent goroutines.
// timeout = per request based timeout for http connections.
// timestamp = the timestamp at which procedure was started.
type rymscrape struct {
	workers      uint
	timeout      uint
	timestamp    string
	reportFolder string
	myclients    []myclient
	jseed        JSeed
}

// postProcess removes duplicate lines in the debug file
func (rym *rymscrape) postProcess() {
	logger.Info("Starting post processing")
	defer logger.Info("Finished post processing")

	removeDuplicatesUnordered := func(elements []string) []string {
		encountered := map[string]bool{}

		// Create a map of all unique elements.
		for v := range elements {
			encountered[elements[v]] = true
		}

		// Place all keys from the map into a slice.
		result := []string{}
		for key := range encountered {
			result = append(result, key)
		}
		return result
	}

	data := readFileIntoList(rym.reportFolder + "/debug")

	data = removeDuplicatesUnordered(data)

	err := os.Remove(rym.reportFolder + "/debug")
	if err != nil {
		panic(err)
	}

	writeToFile(rym.reportFolder+"/debug", strings.Join(data, "\n"))
}

// storeClients stores clients into myclient struct from ./myclients folder
func (dm *rymscrape) storeClients() {
	myclientsDir, err := ioutil.ReadDir("./myclients")
	if err != nil {
		if strings.Contains(err.Error(), "The system cannot find the file specified") {
			logger.Info("No client files detected, progressing without clients...")
			return
		} else {
			panic(err)
		}
	}

	for _, f := range myclientsDir {
		if f.IsDir() {
			continue
		}
		sClient := myclient{}
		sClient.fileName = strings.TrimSuffix(f.Name(), ".txt")
		sClient.data = readFileIntoList("./myclients/" + f.Name())
		sort.Strings(sClient.data)
		dm.myclients = append(dm.myclients, sClient)
	}
	logger.Info("Loaded clients", zap.Int("dm.myclients len", len(dm.myclients)))
}

// start starts the process of collecting links
func (rym *rymscrape) start() {
	var (
		fullLinkList    []string
		episodeLinkList []string
	)

	fullLinkList = rym.getFullList()
	if len(fullLinkList) <= 0 {
		logger.Error("Something wrong with fetching a complete list of brand links")
		return
	}

	if rym.jseed.EpisodeListAcquire.IsTrue {
		np := pool.NewLimited(rym.workers)
		npBatch := np.Batch()

		for _, brandLink := range fullLinkList {
			npBatch.Queue(rym.workerGetEpisodeList(brandLink))
		}
		npBatch.QueueComplete()

		for work := range npBatch.Results() {
			if err := work.Error(); err != nil {
				logger.Error(err.Error())
				continue
			}
			res := work.Value().([]string)
			for _, el := range res {
				if stringInSlice(el, episodeLinkList) {
					continue
				}
				episodeLinkList = append(episodeLinkList, el)
			}
		}
		np.Close()
	} else {
		episodeLinkList = fullLinkList
	}

	if rym.jseed.VideoListAcquire.IsTrue {
		npf := pool.NewLimited(rym.workers)
		npfBatch := npf.Batch()

		for _, episodeLink := range episodeLinkList {
			npfBatch.Queue(rym.workerGetVideoList(episodeLink))
		}
		npfBatch.QueueComplete()

		for work := range npfBatch.Results() {
			if err := work.Error(); err != nil {
				logger.Error(err.Error())
				continue
			}
			result := work.Value().([]reportStructure)
			var data []string
			for _, report := range result {
				data = append(data, fmt.Sprintf(
					"%s\t%s\t%s\t%s",
					report.siteUrl,
					report.pageTitle,
					report.licensor,
					report.cyberlockerLink,
				))
			}
			writeToFile(rym.reportFolder+"/debug", strings.Join(data, "\n"))
		}
		npf.Close()
	}
}

// workerGetEpisodeList is a helper pool function for concurrent routines using "gopkg.in/go-playground/pool.v3" package
func (rym *rymscrape) workerGetEpisodeList(brandLink string) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		if wu.IsCancelled() {
			// return values not used
			return nil, nil
		}

		links, err := rym.getEpisodeList(brandLink)
		if err != nil {
			return nil, err
		}

		return links, nil
	}
}

// workerGetVideoList is a helper pool function for concurrent routines using "gopkg.in/go-playground/pool.v3" package
func (rym *rymscrape) workerGetVideoList(episodeList string) pool.WorkFunc {
	return func(wu pool.WorkUnit) (interface{}, error) {
		if wu.IsCancelled() {
			// return values not used
			return nil, nil
		}

		reports, err := rym.getVideoList(episodeList)
		if err != nil {
			return nil, err
		}

		return reports, nil
	}
}

// getFullList parses through the jseed file and operates based on the commands given to fetch
// the target full brand page links.
func (rym *rymscrape) getFullList() (fullListLinks []string) {
	if len(rym.jseed.FullListLinks) <= 0 {
		logger.Debug("rym.jseed.FullListLinks <= 0")
		return
	}

	for _, fullLink := range rym.jseed.FullListLinks {
		logger.Debug("Discovered fullLink entity", zap.String("fullLink", fullLink))
		fullLink = rym.jseed.SiteProtocol + "://" + rym.jseed.SiteLink + "/" + fullLink

		doc, err := rym.getGoqueryDocument(fullLink)
		if err != nil {
			logger.Error(err.Error())
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
			logger.Error(err.Error())
			continue
		}

		for _, link := range links {
			if stringInSlice(link, fullListLinks) {
				continue
			}
			fullListLinks = append(fullListLinks, link)
		}
	}

	return fullListLinks
}

// getEpisodeList parses through the jseed file and operates based on the commands given to fetch
// the target full brand episode links from the brand page links provided.
func (rym *rymscrape) getEpisodeList(brandLink string) (episodeLinks []string, err error) {
	doc, err := rym.getGoqueryDocument(brandLink)
	if err != nil {
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
	if err != nil {
		return []string{}, err
	}

	return episodeLinks, nil
}

// getVideoList parses through the jseed file and operates based on the commands given to fetch
// the video links from the episode link provided.
func (rym *rymscrape) getVideoList(episodeLink string) (reports []reportStructure, err error) {
	// get page
	pageGoqueryDocument, err := rym.getGoqueryDocument(episodeLink)
	if err != nil {
		return []reportStructure{}, err
	}

	// get title
	pageTitle := pageGoqueryDocument.Find("title").First().Text()

	// get licensor
	licensor := rym.findLicensor(pageTitle)

	switch rym.jseed.VideoListAcquire.Paginate.IsTrue {
	case true:
		var paginatedLinks []string
		p, err := processSeedBlock(
			pageGoqueryDocument,
			rym.jseed.VideoListAcquire.Paginate.LookFor,
			rym.jseed.VideoListAcquire.Paginate.Under,
			rym.jseed.VideoListAcquire.Paginate.Res,
			rym.jseed.SiteProtocol,
			rym.jseed.SiteLink,
		)
		if err != nil {
			return []reportStructure{}, err
		}
		for _, link := range p {
			if stringInSlice(link, paginatedLinks) {
				continue
			}
			paginatedLinks = append(paginatedLinks, link)
		}

		for _, plink := range paginatedLinks {
			doc, err := rym.getGoqueryDocument(plink)
			if err != nil {
				return []reportStructure{}, err
			}

			v, err := processSeedBlock(
				doc,
				rym.jseed.VideoListAcquire.LookFor,
				rym.jseed.VideoListAcquire.Under,
				rym.jseed.VideoListAcquire.Res,
				rym.jseed.SiteProtocol,
				rym.jseed.SiteLink,
			)
			if err != nil {
				return []reportStructure{}, err
			}

			for _, link := range v {
				reports = append(reports, reportStructure{
					siteUrl:         episodeLink,
					licensor:        licensor,
					cyberlockerLink: link,
					pageTitle:       pageTitle,
				})
			}
		}

	case false:
		videoLinks, err := processSeedBlock(
			pageGoqueryDocument,
			rym.jseed.VideoListAcquire.LookFor,
			rym.jseed.VideoListAcquire.Under,
			rym.jseed.VideoListAcquire.Res,
			rym.jseed.SiteProtocol,
			rym.jseed.SiteLink,
		)
		if err != nil {
			return []reportStructure{}, err
		}

		for _, link := range videoLinks {
			reports = append(reports, reportStructure{
				siteUrl:         episodeLink,
				licensor:        licensor,
				cyberlockerLink: link,
				pageTitle:       pageTitle,
			})
		}
	}

	// check deep required
	if rym.jseed.VideoListAcquire.GoDeeper.IsTrue {
		var newReports []reportStructure
		switch rym.jseed.VideoListAcquire.GoDeeper.ByPattern.IsTrue {
		case true:
			for _, report := range reports {
				deepLink, err := rym.getDeepVideoLinkByPattern(report.cyberlockerLink,
					rym.jseed.VideoListAcquire.GoDeeper.ByPattern.PatternStart,
					rym.jseed.VideoListAcquire.GoDeeper.ByPattern.PatternEnd,
				)
				if err != nil {
					logger.Error("Error getting deep link", zap.String("Error", err.Error()))
				} else {
					report.cyberlockerLink = deepLink
				}
				newReports = append(newReports, report)
			}
		case false:
			for _, report := range reports {
				deepLink, err := rym.getDeepVideoLinkByRedirect(report.cyberlockerLink)
				if err != nil {
					logger.Error("Error getting deep link", zap.String("Error", err.Error()))
				} else {
					report.cyberlockerLink = deepLink
				}
				newReports = append(newReports, report)
			}
		}
		reports = newReports
	}

	return reports, nil
}

// findLicensor finds licensor name by analysing all the loaded brand names. The name of the licensor is the filename
// and the brand name is what's contained in the files separated by lines.
func (rym *rymscrape) findLicensor(brandName string) string {
	for _, c := range rym.myclients {
		for _, b := range c.data {
			if strings.Contains(slugify.Slugify(brandName), slugify.Slugify(b)) {
				return c.fileName
			}
		}
	}
	return ""
}

// getGoqueryDocument retrieves the page content in goquery.Document format
func (rym *rymscrape) getGoqueryDocument(link string) (pageGoqueryDocument *goquery.Document, err error) {
	pageRawHTML, _, err := requestGet(link, rym.timeout, false, rym.jseed.SiteSignature)
	if err != nil {
		return pageGoqueryDocument, err
	}

	pageGoqueryDocument, err = goquery.NewDocumentFromReader(bytes.NewReader(pageRawHTML))
	if err != nil {
		return pageGoqueryDocument, err
	}

	return pageGoqueryDocument, nil
}

// getDeepVideoLinkByPattern retrieves deep link by pattern
// if rym.jseed.VideoListAcquire.GoDeeper.ByPattern.isTrue is true
func (rym *rymscrape) getDeepVideoLinkByPattern(link, patternStart, patternEnd string) (deepLink string, err error) {
	return "", nil
}

// getDeepVideoLinkByRedirect retrieves deep link by redirect
// if rym.jseed.VideoListAcquire.GoDeeper.ByRedirect.isTrue is true
func (rym *rymscrape) getDeepVideoLinkByRedirect(link string) (deepLink string, err error) {
	r, _, err := requestGet(link, rym.timeout, true, rym.jseed.SiteSignature)
	if err != nil {
		return "", err
	}

	deepLink = string(r)

	return deepLink, nil
}
