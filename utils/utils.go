package utils

import (
	"encoding/json"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"tui-deck/deck_structs"
)

type Configuration struct {
	User      string `json:"username"`
	Password  string `json:"password"`
	Url       string `json:"url"`
	Color     string `json:"color"`
	ConfigDir string
}

func InitConfingDirectory() (string, error) {
	configDir := getUserDir() + "/.config/tui-deck"
	if !Exists(configDir) {
		err := os.Mkdir(configDir, os.ModePerm)
		if err != nil {
			return "", err
		}
	}
	configFile := configDir + "/config.json"
	if !Exists(configFile) {
		create, err := os.Create(configFile)

		configuration := Configuration{
			User:      "",
			Password:  "",
			Url:       "https://nextcloud.example.com",
			Color:     "#BF40BF",
			ConfigDir: configDir,
		}
		jsonConfig, err := json.Marshal(configuration)
		if err != nil {
			return "", err
		}
		_, err = create.Write(jsonConfig)
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

func Exists(path string) bool {
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

func GetId(name string) int {
	split := strings.Split(name, "-")
	id := 0
	if strings.Contains(split[0], "]") {
		v := strings.Split(strings.Split(split[0], "]")[1], "[")[0]
		_ = v
		id, _ = strconv.Atoi(strings.TrimSpace(strings.ReplaceAll(v, "#", "")))
	} else {
		id, _ = strconv.Atoi(strings.TrimSpace(strings.ReplaceAll(split[0], "#", "")))
	}
	return id
}

func FormatDescription(description string) string {
	return strings.ReplaceAll(description, `\n`, "\n")
}

func BuildLabels(card deck_structs.Card) string {
	var labels = ""
	for i, label := range card.Labels {
		labels = fmt.Sprintf("%s[#%s]%s[white]", labels, label.Color, label.Title)
		if i != len(card.Labels)-1 {
			labels = fmt.Sprintf("%s, ", labels)
		}
	}
	return labels
}

func CleanText(text string) string {
	t1 := strings.ReplaceAll(text, "\"", "\\\"")
	return strings.ReplaceAll(t1, "\n", `\n`)
}

func CommaString(a []string) string {
	res := ""
	for index, j := range a {
		if len(a) == index+1 {
			res = res + fmt.Sprintf("%v", j)
		} else {
			res = res + fmt.Sprintf("%v,", j)
		}

	}
	return res
}
