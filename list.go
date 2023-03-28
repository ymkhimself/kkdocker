package main

import (
	"encoding/json"
	"fmt"
	log "github.com/siruspen/logrus"
	"mydocker/container"
	"os"
	"text/tabwriter"
)

func listContainers() {
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, "")
	dirUrl = dirUrl[:len(dirUrl)-1]
	files, err := os.ReadDir(dirUrl)
	if err != nil {
		log.Errorf("Read dir %s error %v", dirUrl, err)
		return
	}

	var containers []*container.ContainerInfo
	for _, file := range files {
		tmpContainer, err := getContainerInfo(file)
		if err != nil {
			log.Errorf("Get container info error %v", err)
			continue
		}
		containers = append(containers, tmpContainer)
	}
	// tabwriter 用于生成对齐的文本
	w := tabwriter.NewWriter(os.Stdout, 12, 1, 3, ' ', 0)
	fmt.Fprint(w, "ID\tNAME\tPID\tSTATUS\tCOMMAND\tCREATED\n")
	for _, item := range containers {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Id, item.Name, item.Pid, item.Status, item.Command, item.CreatedTime)
	}
	if err := w.Flush(); err != nil {
		log.Errorf("Flush error %v", err)
		return
	}
}

func getContainerInfo(file os.DirEntry) (*container.ContainerInfo, error) {
	containerName := file.Name()
	configDir := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configDir = configDir + container.ConfigName
	content, err := os.ReadFile(configDir)
	if err != nil {
		log.Errorf("Read file %s error %v", configDir, err)
		return nil, err
	}
	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(content, &containerInfo); err != nil {
		log.Errorf("Json unmarshal error %v", err)
		return nil, err
	}
	return &containerInfo, nil
}
