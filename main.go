package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Commands = []cli.Command{
		commandPull,
		commandPush,
		commandPost,
	}
	err := app.Run(os.Args)
	if err != nil {
		logf("error", "%s", err)
		os.Exit(1)
	}
}

func loadConfigFile() (*Config, error) {
	var configFileName string

	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	curFname := filepath.Join(pwd, "blogsync.yaml")
	if _, err := os.Stat(curFname); err == nil {
		configFileName = curFname
	} else {
		home, err := homedir.Dir()
		if err != nil {
			return nil, err
		}
		configFileName = filepath.Join(home, ".config", "blogsync", "config.yaml")
	}
	f, err := os.Open(configFileName)
	if err != nil {
		return nil, err
	}
	return LoadConfig(f)
}

var commandPull = cli.Command{
	Name:  "pull",
	Usage: "Pull entries from remote",
	Action: func(c *cli.Context) error {
		blog := c.Args().First()
		if blog == "" {
			cli.ShowCommandHelp(c, "pull")
			os.Exit(1)
		}

		conf, err := loadConfigFile()
		if err != nil {
			return err
		}
		blogConfig := conf.Get(blog)
		if blogConfig == nil {
			logf("error", "blog not found: %s", blog)
			os.Exit(1)
		}

		b := NewBroker(blogConfig)
		remoteEntries, err := b.FetchRemoteEntries()
		if err != nil {
			return err
		}

		for _, re := range remoteEntries {
			path := b.LocalPath(re)
			_, err := b.StoreFresh(re, path)
			if err != nil {
				return err
			}
		}
		return nil
	},
}

var commandPush = cli.Command{
	Name:  "push",
	Usage: "Push local entries to remote",
	Action: func(c *cli.Context) error {
		path := c.Args().First()
		if path == "" {
			cli.ShowCommandHelp(c, "push")
			os.Exit(1)
		}

		path, err := filepath.Abs(path)
		if err != nil {
			return err
		}

		var blogConfig *BlogConfig

		conf, err := loadConfigFile()
		if err != nil {
			return err
		}
		for remoteRoot := range conf.Blogs {
			bc := conf.Get(remoteRoot)
			localRoot, err := filepath.Abs(filepath.Join(bc.LocalRoot, remoteRoot))
			if err != nil {
				return err
			}

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
		if err != nil {
			return err
		}

		entry, err := entryFromReader(f)
		if err != nil {
			return err
		}

		b.UploadFresh(entry)
		return nil
	},
}

var commandPost = cli.Command{
	Name:  "post",
	Usage: "Post a new entry to remote",
	Flags: []cli.Flag{
		cli.BoolFlag{Name: "draft"},
		cli.StringFlag{Name: "title"},
		cli.StringFlag{Name: "custom-path"},
	},
	Action: func(c *cli.Context) error {
		blog := c.Args().First()
		if blog == "" {
			cli.ShowCommandHelp(c, "post")
			os.Exit(1)
		}

		conf, err := loadConfigFile()
		if err != nil {
			return err
		}
		blogConfig := conf.Get(blog)
		if blogConfig == nil {
			logf("error", "blog not found: %s", blog)
			os.Exit(1)
		}

		entry, err := entryFromReader(os.Stdin)
		if err != nil {
			return err
		}

		if c.Bool("draft") {
			entry.IsDraft = true
		}

		if path := c.String("custom-path"); path != "" {
			entry.CustomPath = path
		}

		if title := c.String("title"); title != "" {
			entry.Title = title
		}

		b := NewBroker(blogConfig)
		err = b.PostEntry(entry)
		if err != nil {
			return err
		}
		return nil
	},
}
