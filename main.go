package main

import (
	"bufio"
	"os"
	"strings"
	"time"

	"github.com/uber-go/zap"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

// VERSION defines the current program release version
const VERSION = "0.0.0.1"

// logger is a declaration for what is instatiated later, we're using zap for speed purposes even though it's not really
// that much of a boost.
var logger zap.Logger

var (
	// Command-line setup
	confWorkers = kingpin.Flag(
		"workers",
		"Number of workers making requests simultaneously and getting the links").
		Default("20").Short('w').Uint()

	confSeedFile = kingpin.Flag(
		"seedfile",
		"Path to the seed file").
		Short('s').String()

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
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		res = append(res, strings.TrimSpace(scanner.Text()))
	}

	err = scanner.Err()
	if err != nil {
		panic(err)
	}

	err = file.Close()
	if err != nil {
		panic(err)
	}

	return res
}

func main() {
	logger = zap.New(
		zap.NewJSONEncoder(),
	)

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
	ryms.timestamp = time.Now().Format("02_01_06-15.04")
	ryms.workers = *confWorkers
	ryms.timeout = *confRequestWaitTimeout
	ryms.jseed = jseed
	ryms.reportFolder = "./_reports_" + jseed.SiteLink
	ryms.storeClients()
	ryms.start()
	ryms.postProcess()

	// generating the report
	generateExcelReport(ryms.timestamp, ryms.reportFolder)

	err := os.Remove(ryms.reportFolder + "/debug")
	if err != nil {
		panic(err)
	}

	logger.Info("Task finished.")
}
