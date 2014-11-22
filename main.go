package main

import (
	"os"

	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Commands = []cli.Command{
		commandPull,
	}

	app.Run(os.Args)
}

var commandPull = cli.Command{
	Name:  "pull",
	Usage: "Pull entries from remote",
	Action: func(c *cli.Context) {
		blog := c.Args().First()
		if blog == "" {
			cli.ShowCommandHelp(c, "pull")
			os.Exit(1)
		}

		blogConfig := config.Get(blog)
		if blogConfig == nil {
			logf("error", "blog not found: %s", blog)
			os.Exit(1)
		}

		b := NewBroker(blogConfig)
		remoteEntries, err := b.FetchRemoteEntries()
		dieIf(err)

		for _, re := range remoteEntries {
			le := b.LocalHalf(re)
			logf("compare", "remote=%s vs local=%s", re.LastModified(), le.LastModified())
			if re.LastModified().After(le.LastModified()) {
				err := b.Download(re, le)
				dieIf(err)
			}
		}
	},
}
