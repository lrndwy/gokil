package gokil-cli-demo

import "github.com/lrndwy/gokil/config"

func LoadSettings() (config.Settings, error) {
	return config.Load(config.Options{})
}
