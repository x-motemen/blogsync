package main

import (
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
)

type Config struct {
	Default *BlogConfig
	Blogs   map[string]*BlogConfig
}

type BlogConfig struct {
	RemoteRoot string `yaml:"-"`
	LocalRoot  string `yaml:"local_root"`
	UserName   string
	Password   string
}

var config *Config = &Config{
	Default: &BlogConfig{
		LocalRoot: "./data",
	},
	Blogs: map[string]*BlogConfig{
		"karimen.hatenablog.com": &BlogConfig{
			RemoteRoot: "karimen.hatenablog.com",
			UserName:   "motemen",
			Password:   "**********",
		},
	},
}

func LoadConfig(r io.Reader) (*Config, error) {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var blogs map[string]*BlogConfig
	err = yaml.Unmarshal(bytes, &blogs)
	if err != nil {
		return nil, err
	}

	config := &Config{
		Default: blogs["default"],
		Blogs:   blogs,
	}

	delete(blogs, "default")
	for key, b := range blogs {
		b.RemoteRoot = key
	}

	return config, nil
}

func (c *Config) Get(remoteRoot string) *BlogConfig {
	conf, ok := c.Blogs[remoteRoot]
	if !ok {
		return nil
	}

	if conf.LocalRoot == "" {
		conf.LocalRoot = config.Default.LocalRoot
	}

	return conf
}
