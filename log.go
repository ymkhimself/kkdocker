package main

import (
	"fmt"
	log "github.com/siruspen/logrus"
	"io"
	"mydocker/container"
	"os"
)

func logContainer(containerName string) {
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	logFilePath := dirUrl + container.ContainerLogFile
	file, err := os.Open(logFilePath)
	defer file.Close()
	if err != nil {
		log.Errorf("Open log file %s error. %v", logFilePath, err)
		return
	}
	content, err := io.ReadAll(file)
	if err != nil {
		log.Errorf("Log container read file %s error. %v", logFilePath, err)
		return
	}
	fmt.Fprintf(os.Stdout, string(content))
}
