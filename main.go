package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/urfave/cli/v2"
)

var errCommandHelp = fmt.Errorf("command help shown")

func main() {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		commandPull,
		commandPush,
		commandPost,
		commandList,
	}
	app.Version = fmt.Sprintf("%s (%s)", version, revision)
	err := app.Run(os.Args)
	if err != nil {
		if err != errCommandHelp {
			logf("error", "%s", err)
		}
		os.Exit(1)
	}
}

func loadSingleConfigFile(fname string) (*config, error) {
	if _, err := os.Stat(fname); err != nil {
		return nil, nil
	}
	f, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return loadConfig(f, fname)
}

func loadConfiguration() (*config, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	var conf *config
	conf, err = loadConfigFiles(pwd)
	if err != nil {
		return nil, err
	}

	var confEnv *config
	confEnv, err = loadConfigFromEnv()
	if err != nil {
		return nil, err
	}
	if conf.Default == nil {
		conf.Default = &blogConfig{}
	}
	if confEnv.Default.Username != "" {
		conf.Default.Username = confEnv.Default.Username
	}
	if confEnv.Default.Password != "" {
		conf.Default.Password = confEnv.Default.Password
	}

	return conf, nil
}

func loadConfigFiles(pwd string) (*config, error) {
	confs := []string{filepath.Join(pwd, "blogsync.yaml")}
	home, err := os.UserHomeDir()
	if err == nil {
		confs = append(confs, filepath.Join(home, ".config", "blogsync", "config.yaml"))
	}
	var conf *config
	for _, confFile := range confs {
		tmpConf, err := loadSingleConfigFile(confFile)
		if err != nil {
			return nil, err
		}
		conf = mergeConfig(conf, tmpConf)
	}
	if conf == nil {
		return nil, fmt.Errorf("no config files found")
	}
	return conf, nil
}

var commandPull = &cli.Command{
	Name:  "pull",
	Usage: "Pull entries from remote",
	Action: func(c *cli.Context) error {
		blog := c.Args().First()
		if blog == "" {
			cli.ShowCommandHelp(c, "pull")
			return errCommandHelp
		}

		conf, err := loadConfiguration()
		if err != nil {
			return err
		}
		blogConfig := conf.Get(blog)
		if blogConfig == nil {
			return fmt.Errorf("blog not found: %s", blog)
		}

		b := newBroker(blogConfig)
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

var commandPush = &cli.Command{
	Name:  "push",
	Usage: "Push local entries to remote",
	Action: func(c *cli.Context) error {
		first := c.Args().First()
		if first == "" {
			cli.ShowCommandHelp(c, "push")
			return errCommandHelp
		}

		conf, err := loadConfiguration()
		if err != nil {
			return err
		}

		for _, path := range c.Args().Slice() {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			entry, err := entryFromReader(f)
			if err != nil {
				return err
			}

			if entry.EditURL == "" {
				// post new entry
				if !filepath.IsAbs(path) {
					var err error
					path, err = filepath.Abs(path)
					if err != nil {
						return err
					}
				}
				bc := conf.detectBlogConfig(path)
				if bc == nil {
					return fmt.Errorf("cannot find blog for %q", path)
				}
				// The entry directory is not always at the top of the localRoot, such as
				// in the case of using subdirectory feature in BlogMedia. Therefore, the
				// relative position from the entry directory is obtained as a custom path as below.
				blogPath, _ := filepath.Rel(bc.localRoot(), path)
				blogPath = "/" + filepath.ToSlash(blogPath)
				stuffs := strings.SplitN(blogPath, "/entry/", 2)
				if len(stuffs) != 2 {
					return fmt.Errorf("%q is not a blog entry", path)
				}
				entry.CustomPath = strings.TrimSuffix(stuffs[1], entryExt)
				b := newBroker(bc)
				err = b.PostEntry(entry, false)
				if err != nil {
					return err
				}
				continue
			}
			blogID, err := entry.blogID()
			if err != nil {
				return err
			}

			bc := conf.Get(blogID)
			if bc == nil {
				return fmt.Errorf("cannot find blog for %s", path)
			}

			_, err = newBroker(bc).UploadFresh(entry)
			if err != nil {
				return err
			}
		}
		return nil
	},
}

var commandPost = &cli.Command{
	Name:  "post",
	Usage: "Post a new entry to remote",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "draft"},
		&cli.StringFlag{Name: "title"},
		&cli.StringFlag{Name: "custom-path"},
		&cli.BoolFlag{Name: "page"},
	},
	Action: func(c *cli.Context) error {
		blog := c.Args().First()
		if blog == "" {
			cli.ShowCommandHelp(c, "post")
			return errCommandHelp
		}

		conf, err := loadConfiguration()
		if err != nil {
			return err
		}
		blogConfig := conf.Get(blog)
		if blogConfig == nil {
			return fmt.Errorf("blog not found: %s", blog)
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

		b := newBroker(blogConfig)
		err = b.PostEntry(entry, c.Bool("page"))
		if err != nil {
			return err
		}
		return nil
	},
}

var commandList = &cli.Command{
	Name:  "list",
	Usage: "List local blogs",
	Action: func(c *cli.Context) error {
		conf, err := loadConfiguration()
		if err != nil {
			return err
		}

		blogs := make([]*struct{ url, fullPath string }, 0, len(conf.Blogs))

		maxURLLen := 0
		for blogID := range conf.Blogs {
			urlLen := len(blogID)
			if urlLen > maxURLLen {
				maxURLLen = urlLen
			}

			blogConfig := conf.Get(blogID)
			var fullPath string
			if blogConfig.OmitDomain == nil || !*blogConfig.OmitDomain {
				fullPath = filepath.Join(blogConfig.LocalRoot, blogConfig.BlogID)
			} else {
				fullPath = blogConfig.LocalRoot
			}
			blogs = append(blogs, &struct{ url, fullPath string }{blogID, fullPath})
		}

		sort.Slice(blogs, func(i, j int) bool { return blogs[i].url < blogs[j].url })

		for _, blog := range blogs {
			del := strings.Repeat(" ", maxURLLen-len(blog.url)+1)
			fmt.Printf("%s%s%s\n", blog.url, del, blog.fullPath)
		}

		return nil
	},
}
