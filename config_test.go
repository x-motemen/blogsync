package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	homedir "github.com/mitchellh/go-homedir"
)

func TestLoadConfigFiles(t *testing.T) {
	orig := homedir.DisableCache
	homedir.DisableCache = true
	defer func() {
		homedir.DisableCache = orig
	}()

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
              default:
                local_root: ./data
              blog1.example.com:
                username: blog1
              blog2.example.com:
                local_root: ./blog2`),
			blogKey: "blog1.example.com",
			expect: blogConfig{
				RemoteRoot: "blog1.example.com",
				LocalRoot:  "./data",
				Username:   "blog1",
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
