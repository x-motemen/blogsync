package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/mitchellh/go-homedir"
)

func main() {
	app := cli.NewApp()
	app.Commands = []cli.Command{
		commandPull,
		commandPush,
	}

	app.Run(os.Args)
}

func loadConfigFile() *Config {
	home, err := homedir.Dir()
	dieIf(err)

	f, err := os.Open(filepath.Join(home, ".config", "blogsync", "config.yaml"))
	dieIf(err)

	conf, err := LoadConfig(f)
	dieIf(err)

	return conf
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

		conf := loadConfigFile()
		blogConfig := conf.Get(blog)
		if blogConfig == nil {
			logf("error", "blog not found: %s", blog)
			os.Exit(1)
		}

		b := NewBroker(blogConfig)
		remoteEntries, err := b.FetchRemoteEntries()
		dieIf(err)

		for _, re := range remoteEntries {
			path := b.LocalPath(re)
			updated, err := b.Mirror(re, path)
			if updated {
				logf("updated", "%s -> %s", re.URL, path)
			}
			dieIf(err)
		}
	},
}

var commandPush = cli.Command{
	Name:  "push",
	Usage: "Push local entries to remote",
	Action: func(c *cli.Context) {
		path := c.Args().First()
		if path == "" {
			cli.ShowCommandHelp(c, "push")
			os.Exit(1)
		}

		path, err := filepath.Abs(path)
		dieIf(err)

		var blogConfig *BlogConfig

		conf := loadConfigFile()
		for remoteRoot := range conf.Blogs {
			bc := conf.Get(remoteRoot)
			localRoot, err := filepath.Abs(filepath.Join(bc.LocalRoot, remoteRoot))
			dieIf(err)

			logf("compare", "%s - %s", path, localRoot)

			if strings.HasPrefix(path, localRoot) {
				blogConfig = bc
				break
			}
		}

		if blogConfig == nil {
			logf("error", "cannot find blog for %s", path)
			os.Exit(1)
		}

		b := NewBroker(blogConfig)

		f, err := os.Open(path)
		dieIf(err)

		entry, err := EntryFromReader(f)
		dieIf(err)

		// TODO リモートのと比較

		logf("entry", "%#v", entry)
		b.Upload(entry)
	},
}
