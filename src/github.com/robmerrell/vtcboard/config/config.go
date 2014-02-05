package config

import (
	"errors"
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"path/filepath"
)

var tomlconfg *toml.TomlTree
var basePath = "resources/configs/"

// LoadConfig takes an environment name and loads that environment's config file.
func LoadConfig(env string) error {
	configPath, err := findConfigPath(".")
	if err != nil {
		return err
	}

	tomlconfg, err = toml.LoadFile(filepath.Join(configPath, basePath, env+".toml"))
	return err
}

// findConfigPath recursively travels up a filepath looking for a resources directory. An
// error is returned in none is eventually found.
func findConfigPath(pathname string) (string, error) {
	absPath, _ := filepath.Abs(pathname)
	if absPath == "/" {
		return "", errors.New("config directory not found")
	}

	files, err := ioutil.ReadDir(pathname)
	if err != nil {
		return "", err
	}

	// look for the resources directory
	for _, f := range files {
		if f.Name() == "resources" {
			return pathname, nil
		}
	}

	return findConfigPath(filepath.Join(pathname, ".."))
}

// String returns a config key as a string
func String(key string) string {
	return tomlconfg.Get(key).(string)
}

// Int returns a config key an int64
func Int(key string) int64 {
	return tomlconfg.Get(key).(int64)
}
