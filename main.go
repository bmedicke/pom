package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

//go:embed config.json
var defaultConfigContent string

//go:embed hooks/default.sh
var defaultHookContent string

//go:embed static/index.html
var defaultIndexContent string

const (
	configfolder = ".config/pom"
	configname   = "config.json"
	staticfolder = "static/"
)

var hookfolder = "hooks/"

// Config is used for unmarshalling.
type Config struct {
	DefaultProject           string `json:"defaultProject"`
	DefaultTask              string `json:"defaultTask"`
	DefaultNote              string `json:"defaultNote"`
	PomodoroDurationMinutes  int    `json:"pomodoroDurationMinutes"`
	BreakDurationMinutes     int    `json:"breakDurationMinutes"`
	LongBreakDurationMinutes int    `json:"longBreakDurationMinutes"`
	LongBreakAfterPomodoros  int    `json:"longBreakAfterPomodoros"`
	WriteTmuxFile            bool   `json:"writeTmuxFile"`
	LogJSON                  bool   `json:"logJSON"`
	HideHeader               bool   `json:"hideHeader"`
	HideFooter               bool   `json:"hideFooter"`
	Server                   string `json:"server"`
	EnableAPI                bool   `json:"enableAPI"`
}

func main() {
	createConfig := flag.Bool(
		"create-config",
		false,
		"create .config/pom config folder with example hooks")

	hookProfile := flag.String(
		"profile",
		"default",
		"select hook profile from ~/.config/pom/hooks/",
	)

	longBreakIn := flag.Int(
		"longbreak-in",
		-1,
		"long break in [c] pomodoros",
	)

	flag.Parse()
	config := getConfig()

	hookfolder = filepath.Join(hookfolder, *hookProfile)

	if *createConfig {
		createConfigFilesAndFolders()
	} else {
		spawnTUI(config, *longBreakIn)
		if config.WriteTmuxFile {
			clearTmuxFile()
		}
	}
}

func createConfigFilesAndFolders() {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Panic(err)
	}

	// create config folders:
	hookpath := filepath.Join(home, configfolder, hookfolder)
	staticpath := filepath.Join(home, configfolder, staticfolder)

	err = os.MkdirAll(hookpath, 0700)
	if err != nil {
		log.Panic(err)
	}

	err = os.MkdirAll(staticpath, 0700)
	if err != nil {
		log.Panic(err)
	}

	// create configfile
	configfile := filepath.Join(home, configfolder, configname)
	_, err = os.Stat(configfile)
	if errors.Is(err, os.ErrNotExist) {
		f, _ := os.Create(configfile)
		defer f.Close()
		f.WriteString(defaultConfigContent)
	}

	defaultHooks := []string{
		"work_start",
		"work_done",
		"break_start",
		"break_done",
		"longbreak_start",
		"longbreak_done",
		"pomodoro_cancelled",
	}

	// create default hook scripts:
	for _, hook := range defaultHooks {
		file := filepath.Join(hookpath, hook)
		_, err := os.Stat(file)
		if errors.Is(err, os.ErrNotExist) {
			f, _ := os.Create(file)
			defer f.Close()
			f.WriteString(defaultHookContent)
			os.Chmod(file, 0700) // make it executable.
		}
	}

	// create index.html:
	file := filepath.Join(staticpath, "index.html")
	_, err = os.Stat(file)
	if errors.Is(err, os.ErrNotExist) {
		f, _ := os.Create(file)
		defer f.Close()
		f.WriteString(defaultIndexContent)
	}

	fmt.Println("the config folder can be found at: ~/.config/pom/")
}

func getConfig() Config {
	var config Config
	home, _ := os.UserHomeDir()
	configpath := filepath.Join(home, configfolder, configname)

	file, _ := os.Open(configpath)
	configJSON, _ := ioutil.ReadAll(file)
	json.Unmarshal([]byte(configJSON), &config)

	return config
}
