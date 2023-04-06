package container

import (
	"fmt"
	log "github.com/siruspen/logrus"
	"os"
	"os/exec"
	"syscall"
)

var (
	RUNNING             string = "running"
	STOP                string = "stop"
	Exit                string = "exited"
	DefaultInfoLocation string = "/var/run/mydocker/%s/"
	ConfigName          string = "config.json"
	ContainerLogFile    string = "container.log"
	RootUrl             string = "/root"
	MntUrl              string = "/root/mnt/%s"
	WriteLayerUrl       string = "/root/writeLayer/%s"
)

type ContainerInfo struct {
	Pid         string `json:"pid"`
	Id          string `json:"id"`
	Name        string `json:"name"`
	Command     string `json:"command"`     // 容器内init进程的运行命令
	CreatedTime string `json:"createdTime"` // 创建时间
	Status      string `json:"status"`      // 状态
	Volume      string `json:"volume"`
	PortMapping []string `json:"portmapping"` //端口映射
}

// NewParentProcess
// 1. /proc/self/exe 调用自己
// 2. args是参数，init是传递给自己的第一个参数，这里会去调用initCommand去初始化一些环境和资源
// 3. 下面的clone参数就是去fork出一个新进程，并且使用namespace隔离新环境和外部环境
func NewParentProcess(tty bool, containerName, volume string, imageName string, envSlice []string) (*exec.Cmd, *os.File) {
	readPipe, writePipe, err := NewPipe()
	if err != nil {
		log.Errorf("New pipe error %v", err)
	}
	cmd := exec.Command("/proc/self/exe", "init")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS |
			syscall.CLONE_NEWNET | syscall.CLONE_NEWIPC,
	}
	if tty { // 如果使用了-ti 参数，就要把当前进程的输入输出绑定到标准输入输出上
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		// 生成容器对应目录下的container.log文件
		dirUrl := fmt.Sprintf(DefaultInfoLocation, containerName)
		if err := os.MkdirAll(dirUrl, 0622); err != nil {
			log.Errorf("New ParentProcess mkdir %s error %v", dirUrl, err)
			return nil, nil
		}
		logFilePath := dirUrl + ContainerLogFile
		logFile, err := os.Create(logFilePath)
		if err != nil {
			log.Errorf("", err)
			return nil, nil
		}
		cmd.Stdout = logFile
	}
	cmd.ExtraFiles = []*os.File{readPipe}
	cmd.Env = append(os.Environ(), envSlice...)
	NewWorkSpace(volume, imageName, containerName)
	cmd.Dir = fmt.Sprintf(MntUrl, containerName)
	return cmd, writePipe
}

func NewPipe() (*os.File, *os.File, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, nil, err
	}
	return read, write, nil
}
