package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

var errCommandHelp = fmt.Errorf("command help shown")

func newApp() *cli.App {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		commandPull,
		commandFetch,
		commandPush,
		commandPost,
		commandList,
		commandRemove,
	}
	app.Version = fmt.Sprintf("%s (%s)", version, revision)
	return app
}

func main() {
	if err := newApp().Run(os.Args); err != nil {
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
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "no-drafts"},
		&cli.BoolFlag{Name: "only-drafts"},
	},
	Action: func(c *cli.Context) error {
		conf, err := loadConfiguration()
		if err != nil {
			return err
		}

		blogs := c.Args().Slice()
		if len(blogs) == 0 {
			blogs = conf.localBlogIDs()
		}
		if len(blogs) == 0 {
			cli.ShowCommandHelp(c, "pull")
			return errCommandHelp
		}

		for _, blog := range blogs {
			blogConfig := conf.Get(blog)
			if blogConfig == nil {
				return fmt.Errorf("blog not found: %s", blog)
			}

			b := newBroker(blogConfig, c.App.Writer)
			remoteEntries, err := b.FetchRemoteEntries(
				!c.Bool("only-drafts"), !c.Bool("no-drafts"))
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
		}
		return nil
	},
}

var commandFetch = &cli.Command{
	Name:  "fetch",
	Usage: "Fetch entries from remote",
	Action: func(c *cli.Context) error {
		first := c.Args().First()
		if first == "" {
			cli.ShowCommandHelp(c, "fetch")
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

			e, err := entryFromReader(f)
			if err != nil {
				return err
			}
			blogID, err := e.blogID()
			if err != nil {
				return err
			}

			bc := conf.Get(blogID)
			if bc == nil {
				return fmt.Errorf("cannot find blog for %s", path)
			}
			b := newBroker(bc, c.App.Writer)
			if _, err := b.StoreFresh(e, path); err != nil {
				return err
			}
		}
		return nil
	},
}

var (
	// 標準フォーマット: 2011/11/07/161845
	defaultBlogPathReg = regexp.MustCompile(`^2[01][0-9]{2}/[01][0-9]/[0-3][0-9]/[0-9]{6}$`)
	// はてなダイアリー風フォーマット: 20111107/1320650325
	hatenaDiaryPathReg = regexp.MustCompile(`^2[01][0-9]{2}[01][0-9][0-3][0-9]/[0-9]{9,12}$`)
	// タイトルフォーマット: 2011/11/07/週末は川に行きました
	titlePathReg = regexp.MustCompile(`^2[01][0-9]{2}/[01][0-9]/[0-3][0-9]/.+$`)
	draftDir     = "_draft/"
)

func isLikelyGivenPath(p string) bool {
	return defaultBlogPathReg.MatchString(p) ||
		hatenaDiaryPathReg.MatchString(p) ||
		titlePathReg.MatchString(p)
}

var commandPush = &cli.Command{
	Name:  "push",
	Usage: "Push local entries to remote",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "publish"},
	},
	Action: func(c *cli.Context) error {
		first := c.Args().First()
		if first == "" {
			cli.ShowCommandHelp(c, "push")
			return errCommandHelp
		}
		publish := c.Bool("publish")

		conf, err := loadConfiguration()
		if err != nil {
			return err
		}

		for _, path := range c.Args().Slice() {
			if !filepath.IsAbs(path) {
				var err error
				path, err = filepath.Abs(path)
				if err != nil {
					return err
				}
			}

			f, err := os.Open(path)
			if err != nil {
				return err
			}
			defer f.Close()

			entry, err := entryFromReader(f)
			if err != nil {
				return err
			}
			if publish && entry.IsDraft {
				entry.IsDraft = false
				// Assume it has been edited and update modtime.
				ti := time.Now()
				entry.LastModified = &ti
			}
			entry.localPath = path

			if entry.EditURL == "" {
				// post new entry
				bc := conf.detectBlogConfig(path)
				if bc == nil {
					return fmt.Errorf("cannot find blog for %q", path)
				}
				// The entry directory is not always at the top of the localRoot, such as
				// in the case of using subdirectory feature in BlogMedia. Therefore, the
				// relative position from the entry directory is obtained as a custom path as below.
				blogPath, _ := filepath.Rel(bc.localRoot(), path)
				blogPath = "/" + filepath.ToSlash(blogPath)
				_, entryPath := extractEntryPath(path)
				if entryPath == "" {
					return fmt.Errorf("%q is not a blog entry", path)
				}
				entry.CustomPath = entryPath
				b := newBroker(bc, c.App.Writer)
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

			blogPath, _ := filepath.Rel(bc.localRoot(), path)
			blogPath = "/" + filepath.ToSlash(blogPath)

			if _, entryPath := extractEntryPath(path); entryPath != "" {
				if !isLikelyGivenPath(entryPath) && !strings.HasPrefix(entryPath, draftDir) {
					entry.CustomPath = entryPath
				}
			}
			_, err = newBroker(bc, c.App.Writer).UploadFresh(entry)
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

		b := newBroker(blogConfig, c.App.Writer)
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

var commandRemove = &cli.Command{
	Name:  "remove",
	Usage: "Remove blog entries",
	Action: func(c *cli.Context) error {
		first := c.Args().First()
		if first == "" {
			cli.ShowCommandHelp(c, "remove")
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

			blogID, err := entry.blogID()
			if err != nil {
				return err
			}

			bc := conf.Get(blogID)
			if bc == nil {
				return fmt.Errorf("cannot find blog for %s", path)
			}

			err = newBroker(bc, c.App.Writer).RemoveEntry(entry, path)
			if err != nil {
				return err
			}
		}
		return nil
	},
}
