package main

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

var homeEnvName = func() string {
	if runtime.GOOS == "windows" {
		return "USERPROFILE"
	}
	return "HOME"
}()

func TestLoadConfigration(t *testing.T) {
	setup := func(t *testing.T, envUsername string, envPassword string, localConf, globalConf *string) (
		string, error) {

		tempdir := t.TempDir()
		for _, envKV := range [][2]string{
			{homeEnvName, tempdir},
			{"BLOGSYNC_USERNAME", envUsername},
			{"BLOGSYNC_PASSWORD", envPassword}} {

			t.Setenv(envKV[0], envKV[1])
		}

		if localConf != nil {
			if runtime.GOOS == "windows" {
				*localConf = strings.ReplaceAll(*localConf, "local_root: /", "local_root: D:/")
			}
			err := os.WriteFile(
				filepath.Join(tempdir, "blogsync.yaml"), []byte(*localConf), 0755)
			if err != nil {
				return "", err
			}
		}

		if globalConf != nil {
			if runtime.GOOS == "windows" {
				*globalConf = strings.ReplaceAll(*globalConf, "local_root: /", "local_root: D:/")
			}
			globalConfFile := filepath.Join(tempdir, ".config", "blogsync", "config.yaml")
			err := os.MkdirAll(filepath.Dir(globalConfFile), 0755)
			if err != nil {
				return "", err
			}
			err = os.WriteFile(globalConfFile, []byte(*globalConf), 0755)
			if err != nil {
				return "", err
			}
		}
		return tempdir, nil
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
			name:      "simple",
			localConf: nil,
			globalConf: pstr(`---
              blog1.example.com:
                username: blog1
                local_root: /data
              blog2.example.com:
                local_root: /blog2`),
			blogKey: "blog1.example.com",
			expect: blogConfig{
				BlogID:    "blog1.example.com",
				LocalRoot: "/data",
				Username:  "blog1",
			},
		},
		{
			name:      "default.local_root",
			localConf: nil,
			globalConf: pstr(`---
              default:
                local_root: /data
              blog1.example.com:
                username: blog1`),
			blogKey: "blog1.example.com",
			expect: blogConfig{
				BlogID:    "blog1.example.com",
				LocalRoot: "/data",
				Username:  "blog1",
			},
		},
		{
			name:      "inherit default config",
			localConf: nil,
			globalConf: pstr(`---
              default:
                username: hoge
                password: fuga
                local_root: /data
                omit_domain: false
              blog2.example.com:
                local_root: /blog2`),
			blogKey: "blog2.example.com",
			expect: blogConfig{
				BlogID:     "blog2.example.com",
				LocalRoot:  "/blog2",
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
                local_root: /data
              blog2.example.com:
                local_root: /blog2`),
			globalConf: nil,
			blogKey:    "blog1.example.com",
			expect: blogConfig{
				BlogID:    "blog1.example.com",
				LocalRoot: "/data",
				Username:  "blog1",
				local:     true,
			},
		},
		{
			name: "merge config and local conf has priority",
			localConf: pstr(`---
              blog1.example.com:
                username: blog1
                local_root: /`),
			globalConf: pstr(`---
              blog1.example.com:
                password: pww
                local_root: ./data`),
			blogKey: "blog1.example.com",
			expect: blogConfig{
				BlogID:    "blog1.example.com",
				LocalRoot: "/",
				Username:  "blog1",
				Password:  "pww",
				local:     true,
			},
		},
		{
			name: "empty configuration",
			localConf: pstr(`---
              default:
                local_root: /ddd
              blog1.example.com:`),
			globalConf: pstr(`---
              default:
                username: mmm
                password: pww
                local_root: /data`),
			blogKey: "blog1.example.com",
			expect: blogConfig{
				BlogID:    "blog1.example.com",
				LocalRoot: "/ddd",
				Username:  "mmm",
				Password:  "pww",
				local:     true,
			},
		},
		{
			name:      "Owner",
			localConf: nil,
			globalConf: pstr(`---
              blog1.example.com:
                username: blog1
                local_root: /data
                owner: sample1
              blog2.example.com:
                local_root: /blog2
                owner: sample2`),
			blogKey: "blog1.example.com",
			expect: blogConfig{
				BlogID:    "blog1.example.com",
				LocalRoot: "/data",
				Username:  "blog1",
				Owner:     "sample1",
			},
		},
		{
			name:        "use system environment and system environment has priority over global conf",
			envUsername: "mmm",
			envPassword: "pww",
			localConf: pstr(`---
              default:
                local_root: /ddd
              blog1.example.com:`),
			globalConf: pstr(`---
              default:
                username: username
                password: password
                local_root: /data`),
			blogKey: "blog1.example.com",
			expect: blogConfig{
				BlogID:    "blog1.example.com",
				LocalRoot: "/ddd",
				Username:  "mmm",
				Password:  "pww",
				local:     true,
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
                local_root: /ddd
              blog1.example.com:`),
			globalConf: pstr(`---
              default:
                local_root: /data`),
			blogKey: "blog1.example.com",
			expect: blogConfig{
				BlogID:    "blog1.example.com",
				LocalRoot: "/ddd",
				Username:  "mmm",
				Password:  "pww",
				local:     true,
			},
		},
		{
			name:        "localConf only, and no system environment",
			envUsername: "",
			envPassword: "",
			localConf: pstr(`---
              blog1.example.com:
                username: blog1
                local_root: /data
              blog2.example.com:
                local_root: /blog2`),
			globalConf: nil,
			blogKey:    "blog1.example.com",
			expect: blogConfig{
				BlogID:    "blog1.example.com",
				LocalRoot: "/data",
				Username:  "blog1",
				local:     true,
			},
		},
		{
			name:        "no default conf, and use auth data from system envrionment",
			envUsername: "mmm",
			envPassword: "pww",
			localConf: pstr(`---
              blog1.example.com:
                local_root: /data`),
			globalConf: nil,
			blogKey:    "blog1.example.com",
			expect: blogConfig{
				BlogID:    "blog1.example.com",
				LocalRoot: "/data",
				Username:  "mmm",
				Password:  "pww",
				local:     true,
			},
		},
		{
			name:      "inherit default config, and no system environment",
			localConf: nil,
			globalConf: pstr(`---
              default:
                username: hoge
                password: fuga
                local_root: /data
                omit_domain: false
              blog2.example.com:
                local_root: /blog2`),
			blogKey: "blog2.example.com",
			expect: blogConfig{
				BlogID:     "blog2.example.com",
				LocalRoot:  "/blog2",
				Username:   "hoge",
				Password:   "fuga",
				OmitDomain: pbool(false),
			},
		},
		{
			name: "config that are only global will have the local flag false",
			localConf: pstr(`---
              blog1.example.com:
                local_root: /blog1`),
			globalConf: pstr(`---
              default:
                username: hoge
                local_root: /data
              blog2.example.com:
                local_root: /blog2`),
			blogKey: "blog2.example.com",
			expect: blogConfig{
				BlogID:    "blog2.example.com",
				LocalRoot: "/blog2",
				Username:  "hoge",
				local:     false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir, err := setup(t, tc.envUsername, tc.envPassword, tc.localConf, tc.globalConf)
			if err != nil {
				t.Fatal(err)
			}
			origPwd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			if err := os.Chdir(dir); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := os.Chdir(origPwd); err != nil {
					t.Fatal(err)
				}
			}()

			conf, err := loadConfiguration()
			if err != nil {
				t.Errorf("error should be nil but: %s", err)
			}
			out := conf.Get(tc.blogKey)

			if runtime.GOOS == "windows" {
				out.LocalRoot = filepath.Clean(out.LocalRoot)
				tc.expect.LocalRoot = filepath.Clean("D:" + tc.expect.LocalRoot)
			}
			if !reflect.DeepEqual(*out, tc.expect) {
				t.Errorf("something went wrong.\n   out: %+v\nexpect: %+v", *out, tc.expect)
			}
		})
	}
}
