package main

import (
	"encoding/json"
	"fmt"
	log "github.com/siruspen/logrus"
	"io/ioutil"
	"mydocker/container"
	"strconv"
	"syscall"
)

/**
获取容器pid
kill
修改容器信息
重新写入容器信息
 */
func stopContainer(containerName string) {
	pid, err := getContainerPidByName(containerName)
	if err != nil {
		log.Errorf("Get container pid by name %s error %v", containerName, err)
		return
	}
	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		log.Errorf("Conver pid from string to int error %v", err)
		return
	}
	if err := syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		log.Errorf("Stop container %s error %v", containerName, err)
		return
	}
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		log.Errorf("Get container %s info err %v", containerName, err)
		return
	}
	containerInfo.Status = container.STOP
	containerInfo.Pid = " "
	newContentBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Json marshal %s error %v", containerName, err)
		return
	}
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirUrl + container.ConfigName
	if err := ioutil.WriteFile(configFilePath, newContentBytes, 0622); err != nil {
		log.Errorf("Write file %s error", configFilePath, err)
	}
}

func getContainerInfoByName(containerName string) (*container.ContainerInfo, error) {
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirUrl + container.ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return nil, err
	}
	return &containerInfo, nil
}
