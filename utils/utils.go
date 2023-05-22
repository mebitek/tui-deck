package utils

import (
	"os"
)

func InitConfingDirectory() string {
	configDir := getUserDir() + "/.config/tui-deck"
	if !exists(configDir) {
		err := os.Mkdir(configDir, os.ModePerm)
		if err != nil {
			return err.Error()
		}
	}
	configFile := configDir + "/config.json"
	if !exists(configFile) {
		create, err := os.Create(configFile)
		if err != nil {
			panic(err.Error())
		}
		if err != nil {
			panic(err.Error())
		}
		return create.Name()
	}
	return configFile

}

func getUserDir() string {
	dirname, err := os.UserHomeDir()
	if err != nil {
		panic(err.Error())
	}
	return dirname
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
