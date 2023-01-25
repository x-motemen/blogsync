package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadConfigFiles(t *testing.T) {
	setup := func(t *testing.T, localConf, globalConf *string) (string, func()) {
		tempdir, err := ioutil.TempDir("", "blogsync-test")
		if err != nil {
			t.Fatal(err)
		}
		origHome := os.Getenv("HOME")
		cleanup := func() {
			os.RemoveAll(tempdir)
			os.Setenv("HOME", origHome)
		}

		if localConf != nil {
			err := ioutil.WriteFile(
				filepath.Join(tempdir, "blogsync.yaml"), []byte(*localConf), 0755)
			if err != nil {
				cleanup()
				t.Fatal(err)
			}
		}

		if globalConf != nil {
			globalConfFile := filepath.Join(tempdir, ".config", "blogsync", "config.yaml")
			err := os.MkdirAll(filepath.Dir(globalConfFile), 0755)
			if err != nil {
				cleanup()
				t.Fatal(err)
			}
			err = ioutil.WriteFile(globalConfFile, []byte(*globalConf), 0755)
			if err != nil {
				cleanup()
				t.Fatal(err)
			}
		}

		err = os.Setenv("HOME", tempdir)
		if err != nil {
			cleanup()
			t.Fatal(err)
		}
		return tempdir, cleanup
	}

	pstr := func(str string) *string {
		return &str
	}
	pbool := func(b bool) *bool {
		return &b
	}
	testCases := []struct {
		name       string
		localConf  *string
		globalConf *string

		blogKey string
		expect  blogConfig
	}{
		{
			name:      "simple",
			localConf: nil,
			globalConf: pstr(`---
              blog1.example.com:
                username: blog1
                local_root: ./data
              blog2.example.com:
                local_root: ./blog2`),
			blogKey: "blog1.example.com",
			expect: blogConfig{
				RemoteRoot: "blog1.example.com",
				LocalRoot:  "./data",
				Username:   "blog1",
			},
		},
		{
			name:      "default.local_root",
			localConf: nil,
			globalConf: pstr(`---
              default:
                local_root: ./data
              blog1.example.com:
                username: blog1`),
			blogKey: "blog1.example.com",
			expect: blogConfig{
				RemoteRoot: "blog1.example.com",
				LocalRoot:  "./data",
				Username:   "blog1",
			},
		},
		{
			name:      "inherit default config",
			localConf: nil,
			globalConf: pstr(`---
              default:
                username: hoge
                password: fuga
                local_root: ./data
                omit_domain: false
              blog2.example.com:
                local_root: ./blog2`),
			blogKey: "blog2.example.com",
			expect: blogConfig{
				RemoteRoot: "blog2.example.com",
				LocalRoot:  "./blog2",
				Username:   "hoge",
				Password:   "fuga",
				OmitDomain: pbool(false),
			},
		},
		{
			name: "localConf only",
			localConf: pstr(`---
              blog1.example.com:
                username: blog1
                local_root: ./data
              blog2.example.com:
                local_root: ./blog2`),
			globalConf: nil,
			blogKey:    "blog1.example.com",
			expect: blogConfig{
				RemoteRoot: "blog1.example.com",
				LocalRoot:  "./data",
				Username:   "blog1",
			},
		},
		{
			name: "merge config and local conf has priority",
			localConf: pstr(`---
              blog1.example.com:
                username: blog1
                local_root: .`),
			globalConf: pstr(`---
              blog1.example.com:
                password: pww
                local_root: ./data`),
			blogKey: "blog1.example.com",
			expect: blogConfig{
				RemoteRoot: "blog1.example.com",
				LocalRoot:  ".",
				Username:   "blog1",
				Password:   "pww",
			},
		},
		{
			name: "empty configuration",
			localConf: pstr(`---
              default:
                local_root: ddd
              blog1.example.com:`),
			globalConf: pstr(`---
              default:
                username: mmm
                password: pww
                local_root: ./data`),
			blogKey: "blog1.example.com",
			expect: blogConfig{
				RemoteRoot: "blog1.example.com",
				LocalRoot:  "ddd",
				Username:   "mmm",
				Password:   "pww",
			},
		},
		{
			name:      "Owner",
			localConf: nil,
			globalConf: pstr(`---
              blog1.example.com:
                username: blog1
                local_root: ./data
                owner: sample1
              blog2.example.com:
                local_root: ./blog2
                owner: sample2`),
			blogKey: "blog1.example.com",
			expect: blogConfig{
				RemoteRoot: "blog1.example.com",
				LocalRoot:  "./data",
				Username:   "blog1",
				Owner:      "sample1",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			workdir, teardown := setup(t, tc.localConf, tc.globalConf)
			defer teardown()
			conf, err := loadConfigFiles(workdir)
			if err != nil {
				t.Errorf("error should be nil but: %s", err)
			}
			out := conf.Get(tc.blogKey)

			if !reflect.DeepEqual(*out, tc.expect) {
				t.Errorf("something went wrong.\n   out: %+v\nexpect: %+v", *out, tc.expect)
			}
		})
	}
}

func TestLoadConfigration(t *testing.T) {
	setup := func(t *testing.T, envUsername string, envPassword string, localConf, globalConf *string) func() {
		tempdir, err := ioutil.TempDir("", "blogsync-test")
		if err != nil {
			t.Fatal(err)
		}
		origHome := os.Getenv("HOME")
		origBlogsyncUsername := os.Getenv("BLOGSYNC_USERNAME")
		origBlogsyncPassword := os.Getenv("BLOGSYNC_PASSWORD")
		origPwd, _ := os.Getwd()
		cleanup := func() {
			os.RemoveAll(tempdir)
			os.Setenv("HOME", origHome)
			os.Setenv("BLOGSYNC_USERNAME", origBlogsyncUsername)
			os.Setenv("BLOGSYNC_PASSWORD", origBlogsyncPassword)
			os.Chdir(origPwd)
		}

		os.Chdir(tempdir)

		if localConf != nil {
			err := ioutil.WriteFile(
				filepath.Join(tempdir, "blogsync.yaml"), []byte(*localConf), 0755)
			if err != nil {
				cleanup()
				t.Fatal(err)
			}
		}

		if globalConf != nil {
			globalConfFile := filepath.Join(tempdir, ".config", "blogsync", "config.yaml")
			err := os.MkdirAll(filepath.Dir(globalConfFile), 0755)
			if err != nil {
				cleanup()
				t.Fatal(err)
			}
			err = ioutil.WriteFile(globalConfFile, []byte(*globalConf), 0755)
			if err != nil {
				cleanup()
				t.Fatal(err)
			}
		}

		err = os.Setenv("HOME", tempdir)
		if err != nil {
			cleanup()
			t.Fatal(err)
		}

		err = os.Setenv("BLOGSYNC_USERNAME", envUsername)
		if err != nil {
			cleanup()
			t.Fatal(err)
		}

		err = os.Setenv("BLOGSYNC_PASSWORD", envPassword)
		if err != nil {
			cleanup()
			t.Fatal(err)
		}

		return cleanup
	}

	pstr := func(str string) *string {
		return &str
	}
	pbool := func(b bool) *bool {
		return &b
	}
	testCases := []struct {
		name        string
		envUsername string
		envPassword string
		localConf   *string
		globalConf  *string

		blogKey string
		expect  blogConfig
	}{
		{
			name:        "use system environment and system environment has priority over global conf",
			envUsername: "mmm",
			envPassword: "pww",
			localConf: pstr(`---
              default:
                local_root: ddd
              blog1.example.com:`),
			globalConf: pstr(`---
              default:
                username: username
                password: password
                local_root: ./data`),
			blogKey: "blog1.example.com",
			expect: blogConfig{
				RemoteRoot: "blog1.example.com",
				LocalRoot:  "ddd",
				Username:   "mmm",
				Password:   "pww",
			},
		},
		{
			name:        "use system environment and system environment has priority over local conf",
			envUsername: "mmm",
			envPassword: "pww",
			localConf: pstr(`---
              default:
                username: username
                password: password
                local_root: ddd
              blog1.example.com:`),
			globalConf: pstr(`---
              default:
                local_root: ./data`),
			blogKey: "blog1.example.com",
			expect: blogConfig{
				RemoteRoot: "blog1.example.com",
				LocalRoot:  "ddd",
				Username:   "mmm",
				Password:   "pww",
			},
		},
		{
			name:        "localConf only, and no system environment",
			envUsername: "",
			envPassword: "",
			localConf: pstr(`---
              blog1.example.com:
                username: blog1
                local_root: ./data
              blog2.example.com:
                local_root: ./blog2`),
			globalConf: nil,
			blogKey:    "blog1.example.com",
			expect: blogConfig{
				RemoteRoot: "blog1.example.com",
				LocalRoot:  "./data",
				Username:   "blog1",
			},
		},
		{
			name:        "no default conf, and use auth data from system envrionment",
			envUsername: "mmm",
			envPassword: "pww",
			localConf: pstr(`---
              blog1.example.com:
                local_root: ./data`),
			globalConf: nil,
			blogKey:    "blog1.example.com",
			expect: blogConfig{
				RemoteRoot: "blog1.example.com",
				LocalRoot:  "./data",
				Username:   "mmm",
				Password:   "pww",
			},
		},
		{
			name:        "inherit default config, and no system environment",
			envUsername: "",
			envPassword: "",
			localConf:   nil,
			globalConf: pstr(`---
              default:
                username: hoge
                password: fuga
                local_root: ./data
                omit_domain: false
              blog2.example.com:
                local_root: ./blog2`),
			blogKey: "blog2.example.com",
			expect: blogConfig{
				RemoteRoot: "blog2.example.com",
				LocalRoot:  "./blog2",
				Username:   "hoge",
				Password:   "fuga",
				OmitDomain: pbool(false),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			teardown := setup(t, tc.envUsername, tc.envPassword, tc.localConf, tc.globalConf)
			defer teardown()
			conf, err := loadConfiguration()
			if err != nil {
				t.Errorf("error should be nil but: %s", err)
			}
			out := conf.Get(tc.blogKey)

			if !reflect.DeepEqual(*out, tc.expect) {
				t.Errorf("something went wrong.\n   out: %+v\nexpect: %+v", *out, tc.expect)
			}
		})
	}
}
