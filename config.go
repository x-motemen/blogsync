package main

import (
	"io"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type config struct {
	Default *blogConfig
	Blogs   map[string]*blogConfig
}

type blogConfig struct {
	RemoteRoot string `yaml:"-"`
	LocalRoot  string `yaml:"local_root"`
	Username   string
	Password   string
}

func LoadConfig(r io.Reader) (*config, error) {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var blogs map[string]*blogConfig
	err = yaml.Unmarshal(bytes, &blogs)
	if err != nil {
		return nil, err
	}

	c := &config{
		Default: blogs["default"],
		Blogs:   blogs,
	}

	delete(blogs, "default")
	for key, b := range blogs {
		b.RemoteRoot = key
	}
	return c, nil
}

func (c *config) Get(remoteRoot string) *blogConfig {
	bc, ok := c.Blogs[remoteRoot]
	if !ok {
		return nil
	}

	if bc.LocalRoot == "" {
		bc.LocalRoot = c.Default.LocalRoot
	}

	return bc
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
	if c1.Default == nil {
		c1.Default = c2.Default
	}
	for k, bc := range c2.Blogs {
		if _, ok := c1.Blogs[k]; !ok {
			c1.Blogs[k] = bc
		}
	}
	return c1
}
