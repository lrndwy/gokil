package demoapi

import "github.com/lrndwy/gokil/config"

func LoadSettings() (config.Settings, error) {
	return config.Load(config.Options{})
}
