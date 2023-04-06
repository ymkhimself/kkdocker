package container

import (
	"fmt"
	log "github.com/siruspen/logrus"
	"os"
	"os/exec"
	"strings"
)

func NewWorkSpace(volume, imageName, containerName string) {
	CreateReadOnlyLayer(imageName)
	CreateWriteLayer(containerName)
	CreateMountPoint(containerName, imageName)
	if volume != "" {
		volumeURLs := volumeUrlExtract(volume)
		length := len(volumeURLs)
		if length == 2 && volumeURLs[0] != "" && volumeURLs[1] != "" {
			MountVolume(volumeURLs, containerName)
			log.Infof("NewWorkSpace volume urls %q", volumeURLs)
		} else {
			log.Infof("Volume parameter input is not correct.")
		}
	}
}

// MountVolume 挂载数据卷
// 1. 读取宿主机文件目录URL，创建宿主机文件目录
// 2. 读取容器挂载点URL，在容器文件系统里创建挂载点
// 3. 把宿主机文件目录挂载到容器挂载点。
func MountVolume(volumeURLs []string, containerName string) {
	// 宿主机文件目录
	parentURL := volumeURLs[0]
	if err := os.MkdirAll(parentURL, 0777); err != nil {
		log.Errorf("Mkdir dir %s error.%v", parentURL, err)
	}
	// 容器文件系统中创建挂载点
	containerUrl := volumeURLs[1]
	mntURL := fmt.Sprintf(MntUrl, containerName)
	containerVolumeURL := mntURL + "/" + containerUrl
	if err := os.Mkdir(containerVolumeURL, 0777); err != nil {
		log.Errorf("Mkdir dir %s error. %v", containerVolumeURL, err)
	}
	// 把宿主机文件目录挂载到容器挂载点上
	dirs := "dirs=" + parentURL
	_, err := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumeURL).CombinedOutput()
	if err != nil {
		log.Errorf("Mount volume failed. %v", err)
	}
}

// 将镜像解压，作为只读层
func CreateReadOnlyLayer(imageName string) {
	unTarFolderUrl := RootUrl + "/" + imageName + "/"
	imageUrl := RootUrl + "/" + imageName + ".tar"
	exist, err := PathExists(unTarFolderUrl)
	if err != nil {
		log.Infof("Fail to judge whether dir %s is exists %v", unTarFolderUrl, err)
		return
	}
	// 如果路径不存在
	if !exist {
		if err := os.Mkdir(unTarFolderUrl, 0622); err != nil {
			log.Errorf("Mkdir dir %s error. %v", unTarFolderUrl, err)
		}
		if _, err := exec.Command("tar", "-xvf", imageUrl, "-C", unTarFolderUrl).CombinedOutput(); err != nil {
			log.Errorf("Untar dit %s err %v", unTarFolderUrl, err)
		}
	}
}

// 创建一个名为writeLayer的文件夹作为容器唯一的可写层
func CreateWriteLayer(containerName string) {
	writeURL := fmt.Sprintf(WriteLayerUrl, containerName)
	if err := os.MkdirAll(writeURL, 0777); err != nil {
		log.Errorf("Mkdir write layer dir %s error. %v", writeURL, err)
	}
}

func CreateMountPoint(containerName string, imageName string) {
	// 创建mnt文件夹作为挂载点
	mntUrl := fmt.Sprintf(MntUrl, containerName)
	if err := os.MkdirAll(mntUrl, 0777); err != nil {
		log.Errorf("Mkdir mountpoint dir %s error. %v", mntUrl, err)
		return
	}

	tmpWriteLayer := fmt.Sprintf(WriteLayerUrl, containerName)
	tmpImageLocation := RootUrl + "/" + imageName
	mntURL := fmt.Sprintf(MntUrl, containerName)
	dirs := "dirs=" + tmpWriteLayer + ":" + tmpImageLocation
	// 把writeLayer目录和busybox目录mount到mnt目录下
	// 第一个是读写目录，后面是只读目录
	_, err := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntURL).CombinedOutput()
	if err != nil {
		log.Errorf("Run command for creating mount point failed %v", err)
	}
}

func DeleteWorkSpace(volume, containerName string) {
	log.Infof("delete workspace")
	if volume != "" {
		volumeURLS := volumeUrlExtract(volume)
		length := len(volumeURLS)
		if length == 2 && volumeURLS[0] != "" && volumeURLS[1] != "" {
			DeleteMountPointWithVolume(volumeURLS, containerName)
		} else {
			DeleteMountPoint(containerName)
		}
	} else {
		DeleteMountPoint(containerName)
	}
	DeleteWriteLayer(containerName)
}

func DeleteMountPointWithVolume(volumeURLS []string, containerName string) {
	mntURL := fmt.Sprintf(MntUrl, containerName)
	containerUrl := mntURL + volumeURLS[1]
	_, err := exec.Command("umount", containerUrl).CombinedOutput()
	if err != nil {
		log.Errorf("Umount volume failed. %v", err)
		return
	}
	_, err = exec.Command("umount", mntURL).CombinedOutput()
	if err != nil {
		log.Errorf("Umount mountpoint failed. %v", err)
		return
	}
	// 删除容器文件系统挂载点
	if err := os.RemoveAll(mntURL); err != nil {
		log.Infof("Remove mountpoint dir %s error %v", mntURL, err)
	}
}

func DeleteWriteLayer(containerName string) {
	writeURL := fmt.Sprintf(WriteLayerUrl, containerName)
	if err := os.RemoveAll(writeURL); err != nil {
		log.Errorf("Remove WriteLayer dir  %s error %v", writeURL, err)
	}
}

func DeleteMountPoint(containerName string) {
	mntURL := fmt.Sprintf(MntUrl, containerName)
	_, err := exec.Command("umount", mntURL).CombinedOutput()
	if err != nil {
		log.Errorf("Unmount %s error %v", mntURL, err)
	}
	if err := os.RemoveAll(mntURL); err != nil {
		log.Errorf("Remove dir %s error %v", mntURL, err)
	}
}

func volumeUrlExtract(volume string) []string {
	var volumeURLs []string
	volumeURLs = strings.Split(volume, ":")
	return volumeURLs
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
