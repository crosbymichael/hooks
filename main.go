package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

var (
	logger      = logrus.New()
	globalFlags = []cli.Flag{
		cli.BoolFlag{Name: "debug", Usage: "enable debug output"},
	}
	globalCommands = []cli.Command{
		serveCommand,
		archiveCommand,
		multiplexCommand,
	}
)

func preload(context *cli.Context) error {
	if context.GlobalBool("debug") {
		logger.Level = logrus.DebugLevel
	}
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
