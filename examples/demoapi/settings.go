package demoapi

import "gokil/config"

func LoadSettings() (config.Settings, error) {
	return config.Load(config.Options{})
}
