package main

import (
	"bufio"
	"os"
	"strings"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

// VERSION defines the current program release version
const VERSION = "0.0.0.1"

var (
	// Command-line setup
	confVerbose = kingpin.Flag(
		"verbose",
		"Toggles verbosity, default is true").
		Default("true").Short('v').Bool()

	confWorkers = kingpin.Flag(
		"workers",
		"Number of workers making requests simultaneously and getting the links").
		Default("2").Short('w').Uint()

	confSeedFile = kingpin.Flag(
		"seedfile",
		"Path to the seed file").
		Default("val.jseed").Short('p').String()

	confRequestWaitTimeout = kingpin.Flag(
		"timeout",
		"Time out for each request after which request is abandoned; Defaults to 30").
		Default("30").Short('t').Uint()
)

// myclient organizes and stores brand names from licensors where
// fileName is the licensor's name and the related data array contains a list of all associated
// brand names.
type myclient struct {
	fileName string
	data     []string
}

// reportStructure is the standard report structure we want things to be arranged in.
type reportStructure struct {
	licensor        string
	siteUrl         string
	pageTitle       string
	cyberlockerLink string
}

// readFileIntoList is a helper function to read a file into a string array
func readFileIntoList(fn string) []string {
	var res []string

	file, err := os.Open(fn)
	handleErrorAndPanic(err)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		res = append(res, strings.TrimSpace(scanner.Text()))
	}

	err = scanner.Err()
	handleErrorAndPanic(err)

	err = file.Close()
	handleErrorAndPanic(err)

	return res
}

func main() {
	//Command-line setup
	kingpin.Version(`
	rymscraper
 *  Contact:
 *  Manish Prakash Singh
 *  contact@kryptodev.com
 *  Skype: kryptodev
	` +
		"\n©rymscraper v" + VERSION + " - removeyourmedia.com, All Rights Reserved.")

	kingpin.Parse()

	// reading the seed file
	jseed := ReadJSeedFile()

	var ryms rymscrape
	ryms.workers = *confWorkers
	ryms.timeout = *confRequestWaitTimeout
	ryms.jseed = jseed
	ryms.reportFolder = "./_reports_" + jseed.SiteLink

	infoLog(ryms.getVideoList("http://www.dramago.com/korean-drama/boys-before-flowers-episode-15"))
}
