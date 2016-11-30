package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fatih/color"
)

var logLocker sync.Mutex

// handleErrorAndPanic is a convenience error handler that panics
func handleErrorAndPanic(err error, msg ...string) {
	if err != nil {
		log.Panicln(err, msg)
	}
}

// debugLog prints given messages if configDebug const is true
func debugLog(msg ...interface{}) {
	logLocker.Lock()
	defer logLocker.Unlock()
	if *confVerbose {
		color.White(fmt.Sprint(time.Now().Format("02_01_06-15.04.05"), "[DEBUG] ->", msg))
	}
}

// errorLog prints error messages
func errorLog(msg ...interface{}) {
	logLocker.Lock()
	defer logLocker.Unlock()
	if *confVerbose {
		color.Red(fmt.Sprint(time.Now().Format("02_01_06-15.04.05"), "[ERROR] ->", msg))
	}
}

// infoLog prints informational messages
func infoLog(msg ...interface{}) {
	logLocker.Lock()
	defer logLocker.Unlock()
	if *confVerbose {
		color.Cyan(fmt.Sprint(time.Now().Format("02_01_06-15.04.05"), "[INFOR] ->", msg))
	}
}

// writeLog prints messages at the time of writing to disk
func writeLog(msg ...interface{}) {
	logLocker.Lock()
	defer logLocker.Unlock()
	if *confVerbose {
		color.Green(fmt.Sprint(time.Now().Format("02_01_06-15.04.05"), "[WRITE] ->", msg))
	}
}
