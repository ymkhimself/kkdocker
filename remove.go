package main

import (
	"fmt"
	log "github.com/siruspen/logrus"
	"mydocker/container"
	"os"
)

/*
*
查找容器信息
容器是否处于停止状态
移除记录容器信息的文件
*/
func removeContainer(containerName string) {
	containerInfo, err := getContainerInfoByName(containerName)
	if err != nil {
		log.Errorf("Get container %s info error %v", containerName)
		return
	}
	if containerInfo.Status != container.STOP {
		log.Errorf("Could't remove running container")
		return
	}
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.RemoveAll(dirUrl); err != nil {
		log.Errorf("Remove file %s error %v", containerName, err)
		return
	}
	container.DeleteWorkSpace(containerInfo.Volume, containerName)
}
