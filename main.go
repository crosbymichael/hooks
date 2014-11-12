package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

var (
	logger      = logrus.New()
	config      *Config
	globalFlags = []cli.Flag{
		cli.BoolFlag{Name: "debug", Usage: "enable debug output"},
		cli.StringFlag{Name: "config,c", Usage: "config file path"},
	}
	globalCommands = []cli.Command{
		githubCommand,
		archiveCommand,
		broadcastCommand,
	}
)

func preload(context *cli.Context) error {
	if context.GlobalBool("debug") {
		logger.Level = logrus.DebugLevel
	}
	config = loadConfig(context.GlobalString("config"))
	return nil
}

func main() {
	app := cli.NewApp()
	app.Name = "hooks"
	app.Usage = "manage github webhooks and events"
	app.Version = "2"
	app.Before = preload
	app.Commands = globalCommands
	app.Flags = globalFlags

	if err := app.Run(os.Args); err != nil {
		logger.Fatal(err)
	}
}
