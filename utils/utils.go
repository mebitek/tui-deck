package utils

import (
	"encoding/json"
	"os"
)

type Configuration struct {
	User     string `json:"username"`
	Password string `json:"password"`
	Url      string `json:"url"`
}

func InitConfingDirectory() (string, error) {
	configDir := getUserDir() + "/.config/tui-deck"
	if !exists(configDir) {
		err := os.Mkdir(configDir, os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	configFile := configDir + "/config.json"
	if !exists(configFile) {
		create, err := os.Create(configFile)
		if err != nil {
			return "", err
		}
		if err != nil {
			return "", err
		}
		return create.Name(), nil
	}
	return configFile, nil

}

func GetConfiguration(configFile string) (Configuration, error) {
	file, _ := os.Open(configFile)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err.Error())
		}
	}(file)
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		return Configuration{}, err
	}
	return configuration, nil
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
