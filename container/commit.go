package container

import (
	"fmt"
	log "github.com/siruspen/logrus"
	"os/exec"
)

func CommitContainer(imageName string) {
	mntUrl := "/root/mnt"
	imageTar := "/root/" + imageName + ".tar"
	fmt.Printf("%s", imageTar)
	if _, err := exec.Command("tar", "-cvf", imageTar, "-C", mntUrl, ".").CombinedOutput(); err != nil {
		log.Errorf("Tar folder %s error %v", mntUrl, err)
	}
}
