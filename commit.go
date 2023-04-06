package main

import (
	"fmt"
	log "github.com/siruspen/logrus"
	"mydocker/container"
	"os/exec"
)

func CommitContainer(containerName string, imageName string) {
	mntUrl := fmt.Sprintf(container.MntUrl, containerName)
	mntUrl += "/"
	imageTar := container.RootUrl + "/" + imageName + ".tar"
		if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntUrl, ".").CombinedOutput(); err != nil {
		log.Errorf("Tar folder %s error %v", mntUrl, err)
	}
}
