// Package config loads configuation objects and provides thin wrappers for interpreting them.
// It reads a search path from the CONFIG_PATH environment variable. All files found along the
// search path are read and cached and are accessible by file name.
//
// Currently, the config package does not support recursive searching; directories found on the
// search path are ignored.
package config

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// DefaultPath is the fallback search path used if EnvVar is not specified.
var DefaultPath string

// EnvVar is the name of the environment variable used to determine the config path.
// If it is provided, it will override the DefaultPath value.
const EnvVar = "CONFIG_PATH"

// Path returns the configuration path value pointed to by EnvVar or DefaultPath.
func Path() string {
	env, ok := os.LookupEnv(EnvVar)
	if !ok {
		return DefaultPath
	}
	return env
}

var (
	pkgOnce sync.Once
	pkgVal  map[string][]byte
	pkgErr  error
)

// Load loads the configuration into memory. After it has been called once, calling
// it again will have no effect.
func Load() error {
	pkgOnce.Do(func() {
		result := map[string][]byte{}
		p := Path()
		log.Printf("config: %s=%s", EnvVar, p)
		ps := filepath.SplitList(p)

		var files []string
		for _, p := range ps {
			fis, err := ioutil.ReadDir(p)
			if err != nil {
				pkgErr = fmt.Errorf("config: error reading directory %q: %w", p, err)
				return
			}
			for _, fi := range fis {
				f := filepath.Join(p, fi.Name())
				if fi.IsDir() {
					continue
				}

				b := filepath.Base(f)
				if result[b] != nil {
					pkgErr = fmt.Errorf("config: multiple config entries with name: %q", b)
					return
				}

				d, err := ioutil.ReadFile(f)
				if err != nil {
					continue
				}

				result[b] = d
				files = append(files, f)
			}
		}
		pkgVal = result

		log.Printf("config: files loaded: %v", strings.Join(files, ", "))
	})

	if pkgErr != nil {
		return fmt.Errorf("config: encountered while loading config: %w", pkgErr)
	}

	return nil
}

// Bytes calls Load() then returns the data for the configuration value named n.
func Bytes(n string) ([]byte, error) {
	err := Load()
	if err != nil {
		return nil, fmt.Errorf("config: failed to get value %q because there was a load error: %w", n, err)
	}

	if v, ok := pkgVal[n]; ok {
		return v, nil
	}

	return nil, os.ErrNotExist
}

// String calls Bytes(n) and converts the result to a string.
func String(n string) (string, error) {
	b, err := Bytes(n)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

type userinfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Userinfo parses configuration value n into a *url.Userinfo struct. It expects
// the input to be a json object with a username and an option password field.
//		{
//			"username": "string",
//			"password": "string"
//		}
func Userinfo(n string) (*url.Userinfo, error) {
	b, err := Bytes(n)
	if err != nil {
		return nil, err
	}

	var ui userinfo
	err = json.Unmarshal(b, &ui)
	if err != nil {
		return nil, fmt.Errorf("config: failed to unmarshal %s into %T: %w", n, new(url.Userinfo), err)
	}

	if ui.Password == "" {
		return url.User(ui.Username), nil
	}

	return url.UserPassword(ui.Username, ui.Password), nil
}

// Url calls url.Parse(String(n))
func Url(n string) (*url.URL, error) {
	s, err := String(n)
	if err != nil {
		return nil, err
	}

	result, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("config: failed to unmarshal %s into %T: %w", n, new(url.URL), err)
	}

	return result, nil
}

// InterfaceJson calls json.Unmarshal() on Bytes(n)
func InterfaceJson(n string, v interface{}) error {
	b, err := Bytes(n)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, v)
	if err != nil {
		return fmt.Errorf("config: failed to unmarshal %s into %T: %w", n, v, err)
	}

	return nil
}

// InterfaceYaml calls yaml.Unmarshal() on Bytes(n)
func InterfaceYaml(n string, v interface{}) error {
	b, err := Bytes(n)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(b, v)
	if err != nil {
		return fmt.Errorf("config: failed to unmarshal %s into %T: %w", n, v, err)
	}

	return nil
}
