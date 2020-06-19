package config

import (
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

// SystemConfig ...
type SystemConfig struct {
	EnableDebug    bool `yaml:"debug_mode"`
	ServerAddr     string `yaml:"server_addr"`
	DefaultProject string `yaml:"default_project"`
	ClientID       string `yaml:"client_id"`
	ClientSecret   string `yaml:"client_secret"`
}

var (
	configDir string
	sysConf SystemConfig
)

func setDefaultParams() {
	sysConf.EnableDebug = false
	sysConf.ServerAddr = "http://localhost:18443"
	sysConf.DefaultProject = "master"
	sysConf.ClientID = "portal"
}

// InitConfig ...
func InitConfig(confDir string) error {
	configDir = confDir
	if configDir == "" {
		// set default path
		conf, err := os.UserConfigDir()
		if err != nil {
			return err
		}
		configDir = filepath.Join(conf, "hekate")
	}

	// mkdir -p configDir
	if _, err := os.Stat(configDir); err != nil {
		err = os.MkdirAll(configDir, 0700)
		if err != nil {
			return err
		}
	}

	// Read Config File
	fname := filepath.Join(configDir, "config.yaml")
	buf, err := ioutil.ReadFile(fname)
	if err != nil {
		// if error is no such file, create file with default params
		if os.IsNotExist(err) {
			setDefaultParams()
			if err := SaveToFile(); err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		if err := yaml.Unmarshal(buf, &sysConf); err != nil {
			return err
		}
	}

	return nil
}

// EnableDebugMode ...
func EnableDebugMode() {
	sysConf.EnableDebug = true
}

// Get ...
func Get() *SystemConfig {
	return &sysConf
}

// SaveToFile ...
func SaveToFile() error {
	fname := filepath.Join(configDir, "config.yaml")
	data, err := yaml.Marshal(sysConf)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fname, data, 0600)
}
