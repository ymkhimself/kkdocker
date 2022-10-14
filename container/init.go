package container

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/siruspen/logrus"
)

// RunContainerInitProcess
// 这里的init函数是在容器内部执行的，也就是说，代码执行到这里之后，容器所在进程已经创建出来了，这时本容器执行的第一个进程。
// 先使用mount挂载proc，以便后续可以通过ps等命令查看当前进程资源情况。
// MS_NOEXEC： 在本文件系统中不允许运行其他程序
// MS_NOSUID： 在本系统中运行程序时，不允许set-user-ID或set-group-ID
// MS_NODEV ： Linux2.4以来，所有mount的系统都会默认设定的参数
// syscall.Exec 是最重要的黑魔法，容器创建后，容器里的第一个程序，即pid为1的进程，是我们指定的前台进程，
// 但是我们的前台进程之前有一个init进程，于是syscall.Exec发挥了作用
// syscall.Exec 调用了内核的execve 系统调用,这个系统调用会将我们指定的进程运行起来，并且将堆栈，数据啥的都覆盖掉，包括PID，这样我们进入容器看到的第一个进程就是这个进程了。
func RunContainerInitProcess() error {
	cmdArray := readUserCommand()
	if cmdArray == nil || len(cmdArray) == 0 {
		return fmt.Errorf("Run container get command error, cmdArray is nil")
	}
	setUpMount()
	path, err := exec.LookPath(cmdArray[0])
	if err != nil {
		log.Errorf("Exec loop path error %v", err)
		return err
	}
	log.Info("Find path %s", path)
	if err := syscall.Exec(path, cmdArray[0:], os.Environ()); err != nil {
		log.Errorf(err.Error())
	}
	return nil
}

// 通过管道读父进程传过来的command
func readUserCommand() []string {
	pipe := os.NewFile(uintptr(3), "pipe")
	msg, err := ioutil.ReadAll(pipe)
	if err != nil {
		log.Errorf("init read pipe error %v", err)
		return nil
	}
	msgStr := string(msg)
	return strings.Split(msgStr, " ")
}

func setUpMount() {
	// 这里要修改一下，不然pivot_root会报错
	if err := syscall.Mount("", "/", "", syscall.MS_PRIVATE|syscall.MS_REC, ""); err != nil {
		logrus.Errorf("mount / fails: %v", err)
		return
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Errorf("Get current location error %v.", err)
	}
	log.Infof("Current location is %s", pwd)
	err = pivotRoot(pwd)
	if err != nil {
		log.Error(err.Error())
	}
	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME,
		"mode=755")
}

// 为了使当前root的老root和新root不在同一个文件系统下，我们把root重新mount一下
// bind mount 是把相同的内容换了一个挂载点的挂载方法
func pivotRoot(root string) error {
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("Mount rootfs to itself error %v", err)
	}
	// 创建rootfs/.pivot_root 存储 old_root
	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0777); err != nil {
		return err
	}
	// pivot_root 到新的rootfs，老的old_root现在挂载rootfs/.pivot_root上
	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root fails %v", err)
	}
	// 修改当前目录到根目录
	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir fails / %v", err)
	}
	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("umount pivot_root dir %v", err)
	}
	// 删除临时文件夹
	return os.Remove(pivotDir)
}
