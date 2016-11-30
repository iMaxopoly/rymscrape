package main

import (
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

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
		Short('p').String()

	confRequestWaitTimeout = kingpin.Flag(
		"timeout",
		"Time out for each request after which request is abandoned; Defaults to 30").
		Default("30").Short('t').Uint()
)

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
}
