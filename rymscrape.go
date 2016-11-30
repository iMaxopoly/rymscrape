package main

import (
	"io/ioutil"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mozillazg/go-slugify"
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

// storeClients stores clients into myclient struct from ./myclients folder
func (dm *rymscrape) storeClients() {
	myclientsDir, err := ioutil.ReadDir("./myclients")
	handleErrorAndPanic(err)

	for _, f := range myclientsDir {
		if !f.IsDir() {
			sClient := myclient{}
			sClient.fileName = strings.TrimSuffix(f.Name(), ".txt")
			sClient.data = readFileIntoList("./myclients/" + f.Name())
			sort.Strings(sClient.data)
			dm.myclients = append(dm.myclients, sClient)
		}
	}
	infoLog("Loaded clients", dm.myclients)
}

// start starts the process of collecting links
func (rym *rymscrape) start() {
	var (
		fullLinkList    []string
		episodeLinkList []string
	)

	fullLinkList = rym.getFullList()

	if len(fullLinkList) <= 0 {
		errorLog("Something wrong with fetching a complete list of brand links")
		return
	}

	np := pool.NewLimited(rym.workers)
	npBatch := np.Batch()

	for _, brandLink := range fullLinkList {
		npBatch.Queue(rym.workerGetEpisodeList(brandLink))
	}
	np.Batch().QueueComplete()

	for work := range npBatch.Results() {
		if err := work.Error(); err != nil {
			errorLog(err)
			continue
		}
		episodeLinkList = append(episodeLinkList, work.Value().([]string)...)
	}
	np.Close()

	np = pool.NewLimited(rym.workers)
	npBatch = np.Batch()

	for _, episodeLink := range episodeLinkList {
		npBatch.Queue(rym.workerGetVideoList(episodeLink))
	}
	np.Batch().QueueComplete()

	for work := range npBatch.Results() {
		if err := work.Error(); err != nil {
			errorLog(err)
			continue
		}
		episodeLinkList = append(episodeLinkList, work.Value().([]reportStructure)...)
	}
	np.Close()
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
			return nil, nil, err
		}

		return reports, nil
	}
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
	doc, err := goquery.NewDocument(episodeLink)
	if err != nil {
		return []reportStructure{}, err
	}

	pageTitle := doc.Find("title").First().Text()
	licensor := func() string {
		for _, c := range rym.myclients {
			for _, b := range c.data {
				if strings.Contains(slugify.Slugify(pageTitle), slugify.Slugify(b)) {
					return c.fileName
				}
			}
		}
		return ""
	}()

	var paginatedLinks []string
	if rym.jseed.VideoListAcquire.Paginate.IsTrue {
		p, err := processSeedBlock(
			doc,
			rym.jseed.VideoListAcquire.Paginate.LookFor,
			rym.jseed.VideoListAcquire.Paginate.Under,
			rym.jseed.VideoListAcquire.Paginate.Res,
			rym.jseed.SiteProtocol,
			rym.jseed.SiteLink,
		)
		if err != nil {
			return []reportStructure{}, err
		}
		paginatedLinks = append(paginatedLinks, p...)

	}

	if len(paginatedLinks) <= 0 {
		videoLinks, err := processSeedBlock(
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

		for _, link := range videoLinks {
			reports = append(reports, reportStructure{
				siteUrl:         episodeLink,
				licensor:        licensor,
				cyberlockerLink: link,
				pageTitle:       pageTitle,
			})
		}

		return reports, nil
	}

	for _, plink := range paginatedLinks {
		doc, err := goquery.NewDocument(plink)
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

	return reports, nil
}
