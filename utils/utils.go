package utils

import (
	"encoding/json"
	"github.com/gdamore/tcell/v2"
	"os"
	"path/filepath"
)

type Configuration struct {
	User     string `json:"username"`
	Password string `json:"password"`
	Url      string `json:"url"`
	Color    string `json:"color"`
}

func InitConfingDirectory() (string, string, error) {
	configDir := getUserDir() + "/.config/tui-deck"
	if !exists(configDir) {
		err := os.Mkdir(configDir, os.ModePerm)
		if err != nil {
			return "", configDir, err
		}
	}
	configFile := configDir + "/config.json"
	if !exists(configFile) {
		create, err := os.Create(configFile)

		configuration := Configuration{
			User:     "",
			Password: "",
			Url:      "https://nextcloud.example.com",
			Color:    "#BF40BF",
		}
		jsonConfig, err := json.Marshal(configuration)
		if err != nil {
			return "", configDir, err
		}
		_, err = create.Write(jsonConfig)
		if err != nil {
			return "", configDir, err
		}
		if err != nil {
			return "", configDir, err
		}
		return create.Name(), configDir, nil
	}
	return configFile, configDir, nil

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

func GetColor(color string) tcell.Color {
	return tcell.GetColor(color)
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

func CreateFile(p string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(p), 0770); err != nil {
		return nil, err
	}
	return os.Create(p)
}
