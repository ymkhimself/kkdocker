package main

import (
	log "github.com/siruspen/logrus"
	"github.com/urfave/cli"
	"os"
)

const usage = "mydocker is a simple container runtime implementation."

func main() {
	app := cli.NewApp()
	app.Name = "mydocker"
	app.Usage = usage
	app.Commands = []cli.Command{
		initCommand,
		runCommand,
		commitCommand,
		listCommand,
		execCommand,
		stopCommand,
		removeCommand,
		logCommand,
	}
	app.Before = func(context *cli.Context) error {
		// 用json作日志代替默认的formatter
		log.SetFormatter(&log.JSONFormatter{})
		log.SetOutput(os.Stdout)
		return nil
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
