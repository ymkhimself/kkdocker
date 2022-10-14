package main

import (
	"mydocker/cgroups"
	"mydocker/cgroups/subsystem"
	"mydocker/container"
	"os"
	"strings"

	log "github.com/siruspen/logrus"
)

func Run(tty bool, comArray []string, res *subsystem.ResourceConfig, volume string) {
	parent, writePipe := container.NewParentProcess(tty, volume) // 创建新的拥有隔离环境的进程
	if parent == nil {
		log.Errorln("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Error(err)
	}
	// 使用my-docker-cgroup 作为新的cgroup name
	// 创建cgroup manager，通过调用set和apply 设置资源限制并使限制在容器上生效
	cgroupManager := cgroups.NewCgroupManager("mydocker-cgroup")
	defer cgroupManager.Destory()
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)
	// 对容器设置完限制后，初始化容器
	sendInitCommand(comArray, writePipe)
	parent.Wait()
	mntURL := "/root/mnt"
	rootURL := "/root"
	container.DeleteWorkSpace(rootURL, mntURL, volume)
	os.Exit(-1)
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
