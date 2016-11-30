package main

import (
	"os"
	"strings"
	"sync"

	"github.com/uber-go/zap"
)

var writeLock sync.Mutex

// writeToFile writes string content to file
func writeToFile(filename, content string) {
	writeLock.Lock()
	defer writeLock.Unlock()

	logger.Info("Writing to", zap.String("filename", filename), zap.String("content", content))
	if strings.TrimSpace(filename) == "" || strings.TrimSpace(content) == "" {
		logger.Error("Writing to", zap.String("filename", filename), zap.String("reason", "failed because content or filename is empty(probably nothing to write)"))
		return
	}

	// ensure path is ready
	folderPath := strings.Split(filename, "/")
	folderPath = folderPath[:len(folderPath)-1]

	folderPathS := strings.Join(folderPath, "/")

	err := os.MkdirAll(folderPathS, 0777)
	if err != nil {
		panic(err)
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}

	_, err = file.WriteString(content + "\n")
	if err != nil {
		panic(err)
	}

	err = file.Close()
	if err != nil {
		panic(err)
	}
}
