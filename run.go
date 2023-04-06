package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"mydocker/cgroups"
	"mydocker/cgroups/subsystem"
	"mydocker/container"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/siruspen/logrus"
)

func Run(tty bool, comArray []string, res *subsystem.ResourceConfig, volume, containerName, imageName string, envSlice []string) {
	containerId := randStringBytes(10)
	if containerName == "" {
		containerName = containerId
	}
	parent, writePipe := container.NewParentProcess(tty, containerName, volume, imageName,envSlice) // 创建新的拥有隔离环境的进程
	if parent == nil {
		log.Errorln("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Error(err)
	}

	// 记录容器信息
	containerName, err := recordContainerInfo(parent.Process.Pid, comArray, containerId, containerName, volume)
	if err != nil {
		log.Errorf("Record container info error %v", err)
		return
	}
	// 使用my-docker-cgroup 作为新的cgroup name
	// 创建cgroup manager，通过调用set和apply 设置资源限制并使限制在容器上生效
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destory()
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)
	// 对容器设置完限制后，初始化容器
	sendInitCommand(comArray, writePipe)
	if tty {
		parent.Wait()
		deleteContainerInfo(containerName)
		container.DeleteWorkSpace(volume, containerName)
	}
}

// 记录容器信息，将信息保存到/var/run/容器名/config.json中
func recordContainerInfo(containerPID int, commandArray []string, containerId, containerName string, volume string) (string, error) {
	// 十位随机数表示容器id
	createTime := time.Now().Format("2006-01-02 15:04:13")
	command := strings.Join(commandArray, "")

	containerInfo := &container.ContainerInfo{
		Name:        containerName,
		Id:          containerId,
		Pid:         strconv.Itoa(containerPID),
		Command:     command,
		CreatedTime: createTime,
		Status:      container.RUNNING,
		Volume:      volume,
	}
	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Record container info error %v", err)
	}
	jsonStr := string(jsonBytes)
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.MkdirAll(dirUrl, 0622); err != nil {
		log.Errorf("Mkdir error %s error %v", dirUrl, err)
		return "", err
	}
	fileName := dirUrl + "/" + container.ConfigName
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		log.Errorf("Create file %s error %v", fileName, err)
		return "", err
	}
	if _, err := file.WriteString(jsonStr); err != nil {
		log.Errorf("File write string error %v", err)
		return "", err
	}
	return containerName, nil
}

func randStringBytes(n int) string {
	letterBytes := "1234567890"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func deleteContainerInfo(containerId string) {
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerId)
	if err := os.RemoveAll(dirUrl); err != nil {
		log.Errorf("Remove dir %s error %v", dirUrl, err)
	}
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
