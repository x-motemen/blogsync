package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

type config struct {
	Default *blogConfig
	Blogs   map[string]*blogConfig
}

func (c *config) detectBlogConfig(fpath string) *blogConfig {
	var retBc *blogConfig
	for blogID := range c.Blogs {
		bc := c.Get(blogID)
		if bc.LocalRoot != "" && strings.HasPrefix(fpath, bc.localRoot()) {
			if retBc == nil || len(bc.localRoot()) > len(retBc.localRoot()) {
				retBc = bc
			}
		}
	}
	return retBc
}

func (c *config) localBlogIDs() []string {
	var ret []string
	for blogID, bc := range c.Blogs {
		if bc.local {
			ret = append(ret, blogID)
		}
	}
	return ret
}

type blogConfig struct {
	BlogID     string `yaml:"-"`
	LocalRoot  string `yaml:"local_root"`
	Username   string
	Password   string
	OmitDomain *bool  `yaml:"omit_domain"`
	Owner      string `yaml:"owner"`
	local      bool
}

func (bc *blogConfig) localRoot() string {
	paths := []string{bc.LocalRoot}
	if bc.OmitDomain == nil || !*bc.OmitDomain {
		paths = append(paths, bc.BlogID)
	}
	return filepath.Join(paths...)
}

func loadConfig(r io.Reader, fpath string) (*config, error) {
	bytes, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if !filepath.IsAbs(fpath) {
		var err error
		fpath, err = filepath.Abs(fpath)
		if err != nil {
			return nil, err
		}
	}
	absDir, fname := filepath.Split(fpath)
	var isLocal = fname == "blogsync.yaml"

	var blogs map[string]*blogConfig
	err = yaml.Unmarshal(bytes, &blogs)
	if err != nil {
		return nil, err
	}

	for key, b := range blogs {
		if b == nil {
			b = &blogConfig{}
		}
		if b.LocalRoot != "" {
			if b.LocalRoot == "~" || strings.HasPrefix(b.LocalRoot, "~/") {
				home, err := os.UserHomeDir()
				if err != nil {
					return nil, err
				}
				b.LocalRoot = strings.Replace(b.LocalRoot, "~", home, 1)
			}
			if !filepath.IsAbs(b.LocalRoot) {
				b.LocalRoot = filepath.Join(absDir, b.LocalRoot)
			}
		}
		if b.BlogID != "default" {
			b.BlogID = key
			b.local = isLocal
		}
		blogs[key] = b
	}

	defaultConf := blogs["default"]
	if defaultConf == nil {
		defaultConf = &blogConfig{}
	}
	delete(blogs, "default")
	return &config{
		Default: defaultConf,
		Blogs:   blogs,
	}, nil
}

func loadConfigFromEnv() (*config, error) {
	return &config{
		Default: &blogConfig{
			Username: os.Getenv("BLOGSYNC_USERNAME"),
			Password: os.Getenv("BLOGSYNC_PASSWORD"),
		},
	}, nil
}

func (c *config) Get(blogID string) *blogConfig {
	bc, ok := c.Blogs[blogID]
	if !ok {
		return nil
	}
	return mergeBlogConfig(bc, c.Default)
}

func mergeBlogConfig(b1, b2 *blogConfig) *blogConfig {
	if b1 == nil {
		if b2 != nil {
			return b2
		}
		b1 = &blogConfig{}
	}
	if b2 == nil {
		return b1
	}
	if b1.LocalRoot == "" {
		b1.LocalRoot = b2.LocalRoot
	}
	if b1.Username == "" {
		b1.Username = b2.Username
	}
	if b1.Password == "" {
		b1.Password = b2.Password
	}
	if b1.OmitDomain == nil {
		b1.OmitDomain = b2.OmitDomain
	}
	if !b1.local {
		b1.local = b2.local
	}
	return b1
}

func mergeConfig(c1, c2 *config) *config {
	if c1 == nil {
		c1 = &config{
			Blogs: make(map[string]*blogConfig),
		}
	}
	if c2 == nil {
		return c1
	}

	c1.Default = mergeBlogConfig(c1.Default, c2.Default)
	for k, bc := range c2.Blogs {
		c1.Blogs[k] = mergeBlogConfig(c1.Blogs[k], bc)
	}
	return c1
}
